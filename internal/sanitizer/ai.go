package sanitizer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type LLMClient interface {
	CheckSensitiveData(ctx context.Context, text string) (bool, error)
}

type AISanitizerRule struct {
	llmClient LLMClient
	cache     map[string]bool
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
}

func NewAISanitizerRule(llmClient LLMClient) *AISanitizerRule {
	return &AISanitizerRule{
		llmClient: llmClient,
		cache:     make(map[string]bool),
		cacheTTL:  24 * time.Hour,
	}
}

func (s *AISanitizerRule) Sanitize(text string) string {
	if s.llmClient == nil {
		return text
	}

	if !s.isSuspicious(text) {
		return text
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	isSensitive, err := s.checkWithCache(ctx, text)
	if err != nil {
		return text
	}

	if isSensitive {
		return s.maskSuspiciousParts(text)
	}

	return text
}

func (s *AISanitizerRule) isSuspicious(text string) bool {
	if len(text) < 10 {
		return false
	}

	if len(text) > 500 {
		return true
	}

	if strings.Contains(strings.ToLower(text), "password") ||
		strings.Contains(strings.ToLower(text), "пароль") ||
		strings.Contains(strings.ToLower(text), "token") ||
		strings.Contains(strings.ToLower(text), "secret") ||
		strings.Contains(strings.ToLower(text), "key") {
		return true
	}

	if len(text) > 30 && s.looksLikeRandomString(text) {
		return true
	}

	return false
}

func (s *AISanitizerRule) looksLikeRandomString(text string) bool {
	alphanumeric := 0
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			alphanumeric++
		}
	}

	ratio := float64(alphanumeric) / float64(len(text))
	return ratio > 0.8 && len(text) > 20
}

func (s *AISanitizerRule) checkWithCache(ctx context.Context, text string) (bool, error) {
	textHash := fmt.Sprintf("%d", len(text))

	s.cacheMu.RLock()
	cached, exists := s.cache[textHash]
	s.cacheMu.RUnlock()

	if exists {
		return cached, nil
	}

	isSensitive, err := s.llmClient.CheckSensitiveData(ctx, text)
	if err != nil {
		return false, err
	}

	s.cacheMu.Lock()
	s.cache[textHash] = isSensitive
	s.cacheMu.Unlock()

	return isSensitive, nil
}

func (s *AISanitizerRule) maskSuspiciousParts(text string) string {
	if len(text) <= 20 {
		return "[FILTERED]"
	}

	if len(text) <= 50 {
		return text[:5] + "..." + text[len(text)-5:] + " [FILTERED]"
	}

	return text[:10] + "...[FILTERED]..." + text[len(text)-10:]
}
