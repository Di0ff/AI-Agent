package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

func formatPrompt(systemMsg, userPrompt string) string {
	return fmt.Sprintf("System: %s\n\nUser: %s", systemMsg, userPrompt)
}

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
