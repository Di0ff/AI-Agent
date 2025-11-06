package llm

import "context"

type Logger interface {
	LogLLMRequest(ctx context.Context, taskID *uint, stepID *uint, role, promptText, responseText, model string, tokensUsed int) error
}

type LLMClient interface {
	PlanAction(ctx context.Context, task string, pageContext string, taskID *uint, stepID *uint) (*StepPlan, error)
}

type StepPlan struct {
	Action     string            `json:"action"`
	Selector   string            `json:"selector"`
	Value      string            `json:"value"`
	Reasoning  string            `json:"reasoning"`
	Parameters map[string]string `json:"parameters"`
}
