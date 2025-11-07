package agent

import (
	"context"
	"strings"
)

type NavigationAgent struct {
	baseAgent *Agent
}

func NewNavigationAgent(baseAgent *Agent) *NavigationAgent {
	return &NavigationAgent{
		baseAgent: baseAgent,
	}
}

func (a *NavigationAgent) CanHandle(ctx context.Context, task string, pageContext string) (float64, error) {
	taskLower := strings.ToLower(task)

	navigationKeywords := []string{
		"открой", "перейди", "найди страницу", "перейти на", "открыть сайт",
		"navigate", "go to", "open", "visit", "find page",
		"поиск", "search", "найти", "ищи",
	}

	for _, keyword := range navigationKeywords {
		if strings.Contains(taskLower, keyword) {
			return 0.9, nil
		}
	}

	return 0.3, nil
}

func (a *NavigationAgent) Execute(ctx context.Context, task string, maxSteps int) error {
	return a.baseAgent.executeTaskString(ctx, task, maxSteps)
}

func (a *NavigationAgent) GetExpertise() []string {
	return []string{"navigation", "search", "page_discovery", "url_handling"}
}

func (a *NavigationAgent) GetType() TaskType {
	return TaskTypeNavigation
}

func (a *NavigationAgent) GetDescription() string {
	return "Specializes in navigation, searching, and page discovery"
}
