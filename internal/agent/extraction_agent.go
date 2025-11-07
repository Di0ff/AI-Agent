package agent

import (
	"context"
	"strings"
)

type ExtractionAgent struct {
	baseAgent *Agent
}

func NewExtractionAgent(baseAgent *Agent) *ExtractionAgent {
	return &ExtractionAgent{
		baseAgent: baseAgent,
	}
}

func (a *ExtractionAgent) CanHandle(ctx context.Context, task string, pageContext string) (float64, error) {
	taskLower := strings.ToLower(task)

	extractionKeywords := []string{
		"найди", "извлеки", "получи информацию", "скопируй", "прочитай",
		"find", "extract", "get information", "copy", "read", "scrape",
		"данные", "информация", "текст", "список",
		"data", "information", "text", "list", "table",
	}

	for _, keyword := range extractionKeywords {
		if strings.Contains(taskLower, keyword) {
			return 0.85, nil
		}
	}

	return 0.2, nil
}

func (a *ExtractionAgent) Execute(ctx context.Context, task string, maxSteps int) error {
	return a.baseAgent.executeTaskString(ctx, task, maxSteps)
}

func (a *ExtractionAgent) GetExpertise() []string {
	return []string{"extraction", "scraping", "data_collection", "reading"}
}

func (a *ExtractionAgent) GetType() TaskType {
	return TaskTypeExtraction
}

func (a *ExtractionAgent) GetDescription() string {
	return "Specializes in data extraction and information gathering"
}
