package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"aiAgent/internal/database"
	"aiAgent/internal/llm"
)

type AgentMemory struct {
	successfulPaths  map[string][]SuccessfulPath
	failurePatterns  map[string]FailurePattern
	siteKnowledge    map[string]SiteInfo
	mu               sync.RWMutex
	repo             *database.TaskRepository
}

type SuccessfulPath struct {
	TaskHash      string
	Steps         []llm.StepPlan
	Strategy      string
	SuccessCount  int
	LastUsed      time.Time
	AverageTime   time.Duration
	Domain        string
}

type FailurePattern struct {
	ErrorType     string
	Action        string
	Selector      string
	Count         int
	LastSeen      time.Time
	Recovery      string
}

type SiteInfo struct {
	Domain         string
	CommonPatterns map[string]string
	FormStructure  []string
	LastVisited    time.Time
}

func NewAgentMemory(repo *database.TaskRepository) *AgentMemory {
	return &AgentMemory{
		successfulPaths: make(map[string][]SuccessfulPath),
		failurePatterns: make(map[string]FailurePattern),
		siteKnowledge:   make(map[string]SiteInfo),
		repo:            repo,
	}
}

func (m *AgentMemory) RecordSuccess(ctx context.Context, task string, steps []llm.StepPlan, strategy string, duration time.Duration, domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskHash := m.hashTask(task)

	path := SuccessfulPath{
		TaskHash:     taskHash,
		Steps:        steps,
		Strategy:     strategy,
		SuccessCount: 1,
		LastUsed:     time.Now(),
		AverageTime:  duration,
		Domain:       domain,
	}

	if existingPaths, ok := m.successfulPaths[taskHash]; ok {
		found := false
		for i, p := range existingPaths {
			if m.pathsAreSimilar(p.Steps, steps) {
				existingPaths[i].SuccessCount++
				existingPaths[i].LastUsed = time.Now()
				existingPaths[i].AverageTime = (existingPaths[i].AverageTime + duration) / 2
				found = true
				break
			}
		}
		if !found {
			m.successfulPaths[taskHash] = append(existingPaths, path)
		}
	} else {
		m.successfulPaths[taskHash] = []SuccessfulPath{path}
	}

	return nil
}

func (m *AgentMemory) RecordFailure(ctx context.Context, action string, selector string, errorMsg string, recovery string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	errorType := m.classifyErrorType(errorMsg)
	key := fmt.Sprintf("%s:%s:%s", errorType, action, selector)

	if pattern, ok := m.failurePatterns[key]; ok {
		pattern.Count++
		pattern.LastSeen = time.Now()
		if recovery != "" {
			pattern.Recovery = recovery
		}
		m.failurePatterns[key] = pattern
	} else {
		m.failurePatterns[key] = FailurePattern{
			ErrorType: errorType,
			Action:    action,
			Selector:  selector,
			Count:     1,
			LastSeen:  time.Now(),
			Recovery:  recovery,
		}
	}

	return nil
}

func (m *AgentMemory) FindSimilarSuccessfulPath(ctx context.Context, task string, domain string) *SuccessfulPath {
	m.mu.RLock()
	defer m.mu.RUnlock()

	taskHash := m.hashTask(task)

	if paths, ok := m.successfulPaths[taskHash]; ok {
		var bestPath *SuccessfulPath
		bestScore := 0

		for _, path := range paths {
			score := path.SuccessCount
			if path.Domain == domain {
				score += 10
			}
			if time.Since(path.LastUsed) < 24*time.Hour {
				score += 5
			}

			if score > bestScore {
				bestScore = score
				pathCopy := path
				bestPath = &pathCopy
			}
		}

		return bestPath
	}

	return nil
}

func (m *AgentMemory) GetFailureRecovery(ctx context.Context, action string, selector string, errorMsg string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errorType := m.classifyErrorType(errorMsg)
	key := fmt.Sprintf("%s:%s:%s", errorType, action, selector)

	if pattern, ok := m.failurePatterns[key]; ok && pattern.Recovery != "" {
		return pattern.Recovery
	}

	generalKey := fmt.Sprintf("%s:%s:", errorType, action)
	for k, pattern := range m.failurePatterns {
		if strings.HasPrefix(k, generalKey) && pattern.Recovery != "" {
			return pattern.Recovery
		}
	}

	return ""
}

func (m *AgentMemory) UpdateSiteKnowledge(ctx context.Context, domain string, patterns map[string]string, forms []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info := SiteInfo{
		Domain:         domain,
		CommonPatterns: patterns,
		FormStructure:  forms,
		LastVisited:    time.Now(),
	}

	m.siteKnowledge[domain] = info
	return nil
}

func (m *AgentMemory) GetSiteKnowledge(ctx context.Context, domain string) *SiteInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if info, ok := m.siteKnowledge[domain]; ok {
		return &info
	}
	return nil
}

func (m *AgentMemory) hashTask(task string) string {
	normalized := strings.ToLower(strings.TrimSpace(task))
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

func (m *AgentMemory) pathsAreSimilar(p1, p2 []llm.StepPlan) bool {
	if len(p1) != len(p2) {
		return false
	}

	matchCount := 0
	for i := range p1 {
		if p1[i].Action == p2[i].Action {
			matchCount++
		}
	}

	return float64(matchCount)/float64(len(p1)) >= 0.8
}

func (m *AgentMemory) classifyErrorType(errorMsg string) string {
	errorLower := strings.ToLower(errorMsg)

	if strings.Contains(errorLower, "timeout") {
		return "timeout"
	}
	if strings.Contains(errorLower, "not found") || strings.Contains(errorLower, "no such element") {
		return "element_not_found"
	}
	if strings.Contains(errorLower, "network") {
		return "network"
	}
	if strings.Contains(errorLower, "permission") || strings.Contains(errorLower, "access denied") {
		return "permission"
	}

	return "unknown"
}

func (m *AgentMemory) SaveToDatabase(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := struct {
		SuccessfulPaths map[string][]SuccessfulPath
		FailurePatterns map[string]FailurePattern
		SiteKnowledge   map[string]SiteInfo
	}{
		SuccessfulPaths: m.successfulPaths,
		FailurePatterns: m.failurePatterns,
		SiteKnowledge:   m.siteKnowledge,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal memory: %w", err)
	}

	_ = jsonData

	return nil
}

func (m *AgentMemory) LoadFromDatabase(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return nil
}
