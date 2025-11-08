// Package agent реализует автономного AI-агента для управления браузером.
// Агент использует LLM (GPT-4o) для планирования действий и специализированных
// подагентов для выполнения задач (навигация, формы, извлечение данных).
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

// UserInputProvider предоставляет интерфейс для взаимодействия с пользователем
// во время выполнения задачи (например, для подтверждения опасных действий).
type UserInputProvider interface {
	// AskUser задает вопрос пользователю и возвращает ответ.
	// Контекст может быть использован для отмены ожидания ввода.
	AskUser(ctx context.Context, question string) (string, error)
}

// Agent представляет автономного агента для управления браузером.
// Агент планирует и выполняет действия, используя LLM и специализированных подагентов.
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
	router            *AgentRouter
	cfg               Config
	memory            *AgentMemory
	circuitBreakers   *CircuitBreakerPool
	reasoningHistory  *llm.ReasoningHistory // История рассуждений для текущей задачи (ReAct pattern)
}

// Config содержит конфигурацию для агента.
type Config struct {
	MaxSteps          int               // Максимальное количество шагов для выполнения задачи
	MaxTokens         int               // Максимальное количество токенов для LLM запросов
	Retries           int               // Количество попыток при ошибках
	RetryDelay        time.Duration     // Задержка между попытками
	UserInputProvider UserInputProvider // Провайдер для взаимодействия с пользователем
	UseSubAgents      bool              // Использовать специализированных подагентов
	ConfidenceMin     float64           // Минимальный уровень уверенности для действий
	UseMultiStep      bool              // Использовать многошаговое планирование
	MultiStepSize     int               // Размер пакета шагов для многошагового планирования
	UseMemory         bool              // Использовать память агента для контекста
}

// ElementPriority определяет приоритет элемента на странице.
type ElementPriority int

const (
	PriorityLow      ElementPriority = iota // Низкий приоритет
	PriorityMedium                          // Средний приоритет
	PriorityHigh                            // Высокий приоритет
	PriorityCritical                        // Критический приоритет (кнопки, ссылки)
)

// PageElement представляет элемент на веб-странице.
type PageElement struct {
	Tag         string          // HTML тег элемента
	Text        string          // Текстовое содержимое
	Selector    string          // CSS селектор для доступа к элементу
	Priority    ElementPriority // Приоритет элемента
	Visible     bool            // Виден ли элемент на странице
	Interactive bool            // Можно ли взаимодействовать с элементом
}
