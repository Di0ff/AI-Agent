package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type MultiStepPlan struct {
	Steps            []StepPlan `json:"steps"`
	OverallStrategy  string     `json:"overall_strategy"`
	FallbackStrategy string     `json:"fallback_strategy"`
	EstimatedSteps   int        `json:"estimated_steps"`
}

func (c *Client) PlanMultiStep(ctx context.Context, task string, pageContext string, maxSteps int, taskID *uint, stepID *uint) (*MultiStepPlan, error) {
	if maxSteps == 0 {
		maxSteps = 5
	}

	category := DetectTaskCategory(task)
	systemMsg := GetSystemPromptForCategory(category)
	fewShot := GetFewShotExamplesForCategory(category)
	guidance := GetTaskSpecificGuidance(category)

	prompt := fmt.Sprintf(`Ты AI-агент, управляющий веб-браузером. Спланируй последовательность действий для выполнения этой задачи.

Задача: %s

Контекст текущей страницы:
%s

%s

%s

Спланируй следующие %d шагов для выполнения задачи. Думай стратегически об оптимальной последовательности.

Доступные действия:
- navigate: перейти по URL
- click: кликнуть на элемент
- type: ввести текст в элемент
- extract_info: извлечь информацию со страницы
- ask_user: спросить пользователя
- complete: задача завершена

Отвечай в формате JSON:
{
  "steps": [
    {
      "action": "action_name",
      "selector": "css_selector (если применимо)",
      "value": "value (если применимо)",
      "reasoning": "почему этот шаг"
    }
  ],
  "overall_strategy": "описание общей стратегии",
  "fallback_strategy": "что делать если шаг не выполнится",
  "estimated_steps": число
}`, task, pageContext, fewShot, guidance, maxSteps)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: systemMsg + "\n\nТы эксперт в планировании многошаговых задач веб-автоматизации. Думай стратегически и планируй заранее. ВСЕГДА отвечай на русском языке.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to plan multi-step: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	content := resp.Choices[0].Message.Content
	var result MultiStepPlan
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse multi-step plan: %w", err)
	}

	if c.logger != nil {
		c.logger.LogLLMRequest(ctx, taskID, stepID, "system", prompt, content, c.model, resp.Usage.TotalTokens)
	}

	return &result, nil
}

func (c *Client) Replan(ctx context.Context, task string, pageContext string, originalPlan *MultiStepPlan, failedStep *StepPlan, errorMessage string, maxSteps int, taskID *uint, stepID *uint) (*MultiStepPlan, error) {
	if maxSteps == 0 {
		maxSteps = 5
	}

	stepsJSON, _ := json.MarshalIndent(originalPlan.Steps, "", "  ")

	prompt := fmt.Sprintf(`Ты AI-агент, управляющий веб-браузером. Оригинальный план провалился, нужно переспланировать.

Задача: %s

Контекст текущей страницы:
%s

Оригинальный план:
%s

Провалившийся шаг:
Действие: %s
Селектор: %s
Значение: %s
Обоснование: %s

Ошибка: %s

Создай новый план, который обойдет эту ошибку. Рассмотри альтернативные подходы.

Спланируй следующие %d шагов. Отвечай в формате JSON:
{
  "steps": [
    {
      "action": "action_name",
      "selector": "css_selector (если применимо)",
      "value": "value (если применимо)",
      "reasoning": "почему этот шаг"
    }
  ],
  "overall_strategy": "описание новой стратегии",
  "fallback_strategy": "что делать если это провалится",
  "estimated_steps": число
}`, task, pageContext, string(stepsJSON), failedStep.Action, failedStep.Selector, failedStep.Value, failedStep.Reasoning, errorMessage, maxSteps)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "Ты эксперт в восстановлении после ошибок и поиске альтернативных подходов к задачам веб-автоматизации. ВСЕГДА отвечай на русском языке.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to replan: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	content := resp.Choices[0].Message.Content
	var result MultiStepPlan
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse replan: %w", err)
	}

	if c.logger != nil {
		c.logger.LogLLMRequest(ctx, taskID, stepID, "system", prompt, content, c.model, resp.Usage.TotalTokens)
	}

	return &result, nil
}
