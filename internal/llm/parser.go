package llm

import (
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

func parseToolCall(toolCall openai.ToolCall) *StepPlan {
	plan := &StepPlan{
		Action:     toolCall.Function.Name,
		Parameters: make(map[string]string),
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		plan.Reasoning = fmt.Sprintf("Ошибка парсинга аргументов: %v", err)
		return plan
	}

	if v, ok := args["selector"].(string); ok {
		plan.Selector = v
	}
	if v, ok := args["value"].(string); ok {
		plan.Value = v
	}
	if v, ok := args["url"].(string); ok {
		plan.Value = v
	}
	if v, ok := args["question"].(string); ok {
		plan.Value = v
	}
	if v, ok := args["reasoning"].(string); ok {
		plan.Reasoning = v
	}

	return plan
}
