package agent

import (
	"context"
	"strings"
)

type FormAgent struct {
	baseAgent *Agent
}

func NewFormAgent(baseAgent *Agent) *FormAgent {
	return &FormAgent{
		baseAgent: baseAgent,
	}
}

func (a *FormAgent) CanHandle(ctx context.Context, task string, pageContext string) (float64, error) {
	taskLower := strings.ToLower(task)
	contextLower := strings.ToLower(pageContext)

	formKeywords := []string{
		"заполни", "введи", "отправь форму", "зарегистрируйся", "войди",
		"fill", "submit", "login", "register", "sign in", "sign up",
		"форма", "form", "input", "поле",
	}

	contextFormIndicators := []string{
		"<form", "input", "textarea", "select", "button[type=\"submit\"]",
		"role=\"form\"", "login", "registration",
	}

	taskScore := 0.0
	for _, keyword := range formKeywords {
		if strings.Contains(taskLower, keyword) {
			taskScore = 0.7
			break
		}
	}

	contextScore := 0.0
	for _, indicator := range contextFormIndicators {
		if strings.Contains(contextLower, indicator) {
			contextScore = 0.3
			break
		}
	}

	return taskScore + contextScore, nil
}

func (a *FormAgent) Execute(ctx context.Context, task string, maxSteps int) error {
	return a.baseAgent.executeTaskString(ctx, task, maxSteps)
}

func (a *FormAgent) GetExpertise() []string {
	return []string{"forms", "input", "validation", "submission", "authentication"}
}

func (a *FormAgent) GetType() TaskType {
	return TaskTypeForm
}

func (a *FormAgent) GetDescription() string {
	return "Specializes in form filling, validation, and submission"
}
