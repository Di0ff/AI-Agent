// Package llm предоставляет интеграцию с OpenAI GPT-4o для планирования действий агента.
// Включает rate limiting, логирование запросов и проверку безопасности действий.
package llm

import "context"

// Logger определяет интерфейс для логирования LLM запросов.
type Logger interface {
	// LogLLMRequest сохраняет информацию о запросе к LLM в базу данных.
	LogLLMRequest(ctx context.Context, taskID *uint, stepID *uint, role, promptText, responseText, model string, tokensUsed int) error
}

// LLMClient определяет интерфейс для взаимодействия с LLM.
// Все методы поддерживают контекст для отмены и логирование запросов.
type LLMClient interface {
	// Reason выполняет фазу явного рассуждения перед планированием действия (ReAct pattern).
	// Агент анализирует ситуацию, вырабатывает стратегию и оценивает уверенность.
	Reason(ctx context.Context, task string, pageContext string, history *ReasoningHistory, taskID *uint, stepID *uint) (*ReasoningStep, error)

	// ReasonWithContext выполняет reasoning с учетом релевантных паттернов из памяти агента.
	ReasonWithContext(ctx context.Context, task string, pageContext string, history *ReasoningHistory, memoryContext string, taskID *uint, stepID *uint) (*ReasoningStep, error)

	// PlanActionWithReasoning планирует действие с учетом reasoning context (ReAct pattern).
	// Reasoning направляет планирование - агент планирует на основе выработанной стратегии.
	PlanActionWithReasoning(ctx context.Context, task string, pageContext string, reasoning *ReasoningStep, taskID *uint, stepID *uint) (*StepPlan, error)

	// PlanAction планирует следующее действие на основе задачи и контекста страницы (legacy).
	// Для новой архитектуры с ReAct pattern используй PlanActionWithReasoning.
	PlanAction(ctx context.Context, task string, pageContext string, taskID *uint, stepID *uint) (*StepPlan, error)

	// CheckDangerousAction проверяет является ли действие потенциально опасным.
	CheckDangerousAction(ctx context.Context, action, selector, value, reasoning string) (bool, string, error)

	// PlanMultiStep создает план из нескольких шагов для выполнения сложной задачи.
	PlanMultiStep(ctx context.Context, task string, pageContext string, maxSteps int, taskID *uint, stepID *uint) (*MultiStepPlan, error)

	// Replan пересоздает план после ошибки выполнения шага.
	Replan(ctx context.Context, task string, pageContext string, originalPlan *MultiStepPlan, failedStep *StepPlan, errorMessage string, maxSteps int, taskID *uint, stepID *uint) (*MultiStepPlan, error)
}

// StepPlan представляет план одного шага действия.
type StepPlan struct {
	Action     string            `json:"action"`     // Тип действия (navigate, click, type и т.д.)
	Selector   string            `json:"selector"`   // CSS селектор элемента
	Value      string            `json:"value"`      // Значение для ввода (для type)
	Reasoning  string            `json:"reasoning"`  // Обоснование действия
	Parameters map[string]string `json:"parameters"` // Дополнительные параметры
}
