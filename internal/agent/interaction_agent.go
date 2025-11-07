package agent

import (
	"context"
	"strings"
)

type InteractionAgent struct {
	baseAgent *Agent
}

func NewInteractionAgent(baseAgent *Agent) *InteractionAgent {
	return &InteractionAgent{
		baseAgent: baseAgent,
	}
}

func (a *InteractionAgent) CanHandle(ctx context.Context, task string, pageContext string) (float64, error) {
	taskLower := strings.ToLower(task)

	interactionKeywords := []string{
		"кликни", "нажми", "выбери", "прокрути", "наведи",
		"click", "press", "select", "scroll", "hover", "drag", "drop",
		"кнопка", "ссылка", "меню",
		"button", "link", "menu", "tab",
	}

	for _, keyword := range interactionKeywords {
		if strings.Contains(taskLower, keyword) {
			return 0.8, nil
		}
	}

	return 0.4, nil
}

func (a *InteractionAgent) Execute(ctx context.Context, task string, maxSteps int) error {
	return a.baseAgent.executeTaskString(ctx, task, maxSteps)
}

func (a *InteractionAgent) GetExpertise() []string {
	return []string{"clicking", "scrolling", "hovering", "interaction", "ui_manipulation"}
}

func (a *InteractionAgent) GetType() TaskType {
	return TaskTypeInteraction
}

func (a *InteractionAgent) GetDescription() string {
	return "Specializes in UI interactions like clicking, scrolling, and hovering"
}
