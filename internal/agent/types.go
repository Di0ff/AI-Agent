package agent

import (
	"aiAgent/internal/browser"
	"aiAgent/internal/database"
	"aiAgent/internal/llm"
	"aiAgent/internal/logger"
	"aiAgent/internal/sanitizer"
	"context"
	"time"
)

type UserInputProvider interface {
	AskUser(ctx context.Context, question string) (string, error)
}

type Agent struct {
	browser           browser.Browser
	llmClient         llm.LLMClient
	repo              *database.TaskRepository
	log               *logger.Zap
	maxSteps          int
	maxTokens         int
	retries           int
	retryDelay        time.Duration
	userInputProvider UserInputProvider
	securityChecker   *SecurityChecker
	sanitizer         *sanitizer.DataSanitizer
}

type Config struct {
	MaxSteps          int
	MaxTokens         int
	Retries           int
	RetryDelay        time.Duration
	UserInputProvider UserInputProvider
}

type ElementPriority int

const (
	PriorityLow ElementPriority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

type PageElement struct {
	Tag         string
	Text        string
	Selector    string
	Priority    ElementPriority
	Visible     bool
	Interactive bool
}
