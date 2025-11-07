package llm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter реализует token bucket algorithm для ограничения частоты запросов
type RateLimiter struct {
	requestsPerMinute int
	tokensPerHour     int

	// Request rate limiting (RPM)
	requestTokens    int
	requestCapacity  int
	requestMu        sync.Mutex
	requestLastCheck time.Time

	// Token rate limiting (TPH)
	tokenBudget    int
	tokenCapacity  int
	tokenMu        sync.Mutex
	tokenLastCheck time.Time
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(requestsPerMinute, tokensPerHour int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60 // дефолт: 60 запросов в минуту
	}
	if tokensPerHour <= 0 {
		tokensPerHour = 90000 // дефолт: 90k токенов в час (GPT-4 tier 1)
	}

	now := time.Now()
	return &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		tokensPerHour:     tokensPerHour,
		requestTokens:     requestsPerMinute,
		requestCapacity:   requestsPerMinute,
		requestLastCheck:  now,
		tokenBudget:       tokensPerHour,
		tokenCapacity:     tokensPerHour,
		tokenLastCheck:    now,
	}
}

// refillRequestTokens пополняет токены запросов
func (rl *RateLimiter) refillRequestTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.requestLastCheck)

	// Пополняем токены пропорционально прошедшему времени
	tokensToAdd := int(elapsed.Minutes() * float64(rl.requestsPerMinute))
	rl.requestTokens += tokensToAdd

	if rl.requestTokens > rl.requestCapacity {
		rl.requestTokens = rl.requestCapacity
	}

	rl.requestLastCheck = now
}

// refillTokenBudget пополняет бюджет токенов
func (rl *RateLimiter) refillTokenBudget() {
	now := time.Now()
	elapsed := now.Sub(rl.tokenLastCheck)

	// Пополняем бюджет пропорционально прошедшему времени
	tokensToAdd := int(elapsed.Hours() * float64(rl.tokensPerHour))
	rl.tokenBudget += tokensToAdd

	if rl.tokenBudget > rl.tokenCapacity {
		rl.tokenBudget = rl.tokenCapacity
	}

	rl.tokenLastCheck = now
}

// AllowRequest проверяет, можно ли выполнить запрос
func (rl *RateLimiter) AllowRequest(ctx context.Context) error {
	rl.requestMu.Lock()
	defer rl.requestMu.Unlock()

	rl.refillRequestTokens()

	if rl.requestTokens <= 0 {
		waitTime := time.Minute / time.Duration(rl.requestsPerMinute)
		return fmt.Errorf("превышен лимит запросов (%d RPM), повторите через %v", rl.requestsPerMinute, waitTime)
	}

	rl.requestTokens--
	return nil
}

// AllowTokens проверяет, можно ли использовать указанное количество токенов
func (rl *RateLimiter) AllowTokens(ctx context.Context, tokens int) error {
	rl.tokenMu.Lock()
	defer rl.tokenMu.Unlock()

	rl.refillTokenBudget()

	if rl.tokenBudget < tokens {
		waitTime := time.Hour / time.Duration(rl.tokensPerHour/tokens)
		return fmt.Errorf("превышен лимит токенов (%d TPH), недостаточно токенов (%d требуется, %d доступно), повторите через %v",
			rl.tokensPerHour, tokens, rl.tokenBudget, waitTime)
	}

	rl.tokenBudget -= tokens
	return nil
}

// ConsumeTokens списывает токены после успешного запроса
func (rl *RateLimiter) ConsumeTokens(tokens int) {
	rl.tokenMu.Lock()
	defer rl.tokenMu.Unlock()

	rl.tokenBudget -= tokens
	if rl.tokenBudget < 0 {
		rl.tokenBudget = 0
	}
}

// GetStats возвращает текущую статистику лимитера
func (rl *RateLimiter) GetStats() (requestsAvailable int, tokensAvailable int) {
	rl.requestMu.Lock()
	rl.refillRequestTokens()
	requestsAvailable = rl.requestTokens
	rl.requestMu.Unlock()

	rl.tokenMu.Lock()
	rl.refillTokenBudget()
	tokensAvailable = rl.tokenBudget
	rl.tokenMu.Unlock()

	return
}
