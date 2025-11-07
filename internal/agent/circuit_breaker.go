package agent

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	state        CircuitState
	failures     int
	lastFailure  time.Time
	mu           sync.RWMutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	if maxFailures == 0 {
		maxFailures = 5
	}
	if resetTimeout == 0 {
		resetTimeout = 30 * time.Second
	}

	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
		failures:     0,
	}
}

func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
	cb.mu.Lock()

	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.failures = 0
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker is open")
		}
	}

	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}

		return err
	}

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
	}
	cb.failures = 0

	return nil
}

func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
}

func RetryWithExponentialBackoff(ctx context.Context, maxRetries int, baseDelay time.Duration, fn func() error) error {
	if maxRetries == 0 {
		maxRetries = 3
	}
	if baseDelay == 0 {
		baseDelay = 1 * time.Second
	}

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
			maxDelay := 30 * time.Second
			if delay > maxDelay {
				delay = maxDelay
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		actionErr := classifyError("", err)
		if actionErr.Type == ErrorTypeCritical {
			return err
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

type CircuitBreakerPool struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
}

func NewCircuitBreakerPool() *CircuitBreakerPool {
	return &CircuitBreakerPool{
		breakers: make(map[string]*CircuitBreaker),
	}
}

func (pool *CircuitBreakerPool) GetBreaker(key string) *CircuitBreaker {
	pool.mu.RLock()
	if breaker, ok := pool.breakers[key]; ok {
		pool.mu.RUnlock()
		return breaker
	}
	pool.mu.RUnlock()

	pool.mu.Lock()
	defer pool.mu.Unlock()

	if breaker, ok := pool.breakers[key]; ok {
		return breaker
	}

	breaker := NewCircuitBreaker(5, 30*time.Second)
	pool.breakers[key] = breaker
	return breaker
}

func (pool *CircuitBreakerPool) ResetAll() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for _, breaker := range pool.breakers {
		breaker.Reset()
	}
}
