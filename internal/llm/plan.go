package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

func formatPrompt(systemMsg, userPrompt string) string {
	return fmt.Sprintf("System: %s\n\nUser: %s", systemMsg, userPrompt)
}

// PlanActionWithReasoning планирует действие С УЧЕТОМ предыдущего reasoning.
// Это новый метод для работы с ReAct pattern - reasoning направляет планирование.
func (c *Client) PlanActionWithReasoning(ctx context.Context, task string, pageContext string, reasoning *ReasoningStep, taskID *uint, stepID *uint) (*StepPlan, error) {
	tools := getTools()

	// Используем минимальный unified prompt (без категорий)
	systemMsg := GetSystemPromptForCategory(CategoryGeneral)

	// Формируем prompt с reasoning context
	prompt := fmt.Sprintf(`Текущая задача: %s

Контекст страницы:
%s`, task, pageContext)

	// Добавляем reasoning context если доступен
	if reasoning != nil {
		prompt += fmt.Sprintf(`

Результат reasoning фазы:
- Наблюдение: %s
- Анализ: %s
- Стратегия: %s
- Уверенность: %.2f

Планируй действие на основе выработанной стратегии.`,
			reasoning.Observation,
			reasoning.Analysis,
			reasoning.Strategy,
			reasoning.Confidence)
	}

	prompt += "\n\nОпредели КОНКРЕТНОЕ следующее действие для выполнения задачи. Используй tool calling."

	resp, err := c.createChatCompletionWithRateLimit(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMsg,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Tools: tools,
	})

	if err != nil {
		if c.logger != nil {
			fullPrompt := formatPrompt(systemMsg, prompt)
			sanitizedPrompt := c.sanitizer.Sanitize(fullPrompt)
			sanitizedError := c.sanitizer.Sanitize(err.Error())
			_ = c.logger.LogLLMRequest(ctx, taskID, stepID, "planning_error", sanitizedPrompt, sanitizedError, c.model, 0)
		}
		return nil, fmt.Errorf("ошибка запроса к OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("пустой ответ от OpenAI")
	}

	choice := resp.Choices[0]
	var responseText string
	var toolCall *openai.ToolCall

	if len(choice.Message.ToolCalls) > 0 {
		toolCall = &choice.Message.ToolCalls[0]
		responseText = fmt.Sprintf("Tool call: %s(%s)", toolCall.Function.Name, toolCall.Function.Arguments)
	} else {
		responseText = choice.Message.Content
	}

	if c.logger != nil {
		fullPrompt := formatPrompt(systemMsg, prompt)
		sanitizedPrompt := c.sanitizer.Sanitize(fullPrompt)
		sanitizedResponse := c.sanitizer.Sanitize(responseText)
		tokensUsed := resp.Usage.TotalTokens
		_ = c.logger.LogLLMRequest(ctx, taskID, stepID, "planning", sanitizedPrompt, sanitizedResponse, c.model, tokensUsed)
	}

	if toolCall != nil {
		return parseToolCall(*toolCall), nil
	}

	return &StepPlan{
		Action:    "complete",
		Reasoning: choice.Message.Content,
	}, nil
}

// PlanAction планирует действие БЕЗ reasoning context (legacy метод).
// Для новой архитектуры используй PlanActionWithReasoning.
// Сохранено для обратной совместимости с кодом который не использует reasoning.
func (c *Client) PlanAction(ctx context.Context, task string, pageContext string, taskID *uint, stepID *uint) (*StepPlan, error) {
	tools := getTools()

	category := DetectTaskCategory(task)
	systemMsg := GetSystemPromptForCategory(category)
	fewShot := GetFewShotExamplesForCategory(category)
	guidance := GetTaskSpecificGuidance(category)

	prompt := fmt.Sprintf(`Текущая задача: %s

Контекст страницы:
%s

%s

%s

Определи следующее действие для выполнения задачи. Используй доступные инструменты для взаимодействия с браузером.`, task, pageContext, fewShot, guidance)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMsg,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Tools: tools,
	})

	if err != nil {
		if c.logger != nil {
			fullPrompt := formatPrompt(systemMsg, prompt)
			sanitizedPrompt := c.sanitizer.Sanitize(fullPrompt)
			sanitizedError := c.sanitizer.Sanitize(err.Error())
			_ = c.logger.LogLLMRequest(ctx, taskID, stepID, "error", sanitizedPrompt, sanitizedError, c.model, 0)
		}
		return nil, fmt.Errorf("ошибка запроса к OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("пустой ответ от OpenAI")
	}

	choice := resp.Choices[0]
	var responseText string
	var toolCall *openai.ToolCall

	if len(choice.Message.ToolCalls) > 0 {
		toolCall = &choice.Message.ToolCalls[0]
		responseText = fmt.Sprintf("Tool call: %s(%s)", toolCall.Function.Name, toolCall.Function.Arguments)
	} else {
		responseText = choice.Message.Content
	}

	if c.logger != nil {
		fullPrompt := formatPrompt(systemMsg, prompt)
		sanitizedPrompt := c.sanitizer.Sanitize(fullPrompt)
		sanitizedResponse := c.sanitizer.Sanitize(responseText)
		tokensUsed := resp.Usage.TotalTokens
		_ = c.logger.LogLLMRequest(ctx, taskID, stepID, "assistant", sanitizedPrompt, sanitizedResponse, c.model, tokensUsed)
	}

	if toolCall != nil {
		return parseToolCall(*toolCall), nil
	}

	return &StepPlan{
		Action:    "complete",
		Reasoning: choice.Message.Content,
	}, nil
}
