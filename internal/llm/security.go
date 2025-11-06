package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type SecurityCheckResult struct {
	IsDangerous bool   `json:"is_dangerous"`
	Reason      string `json:"reason"`
	Message     string `json:"message"`
}

func (c *Client) CheckDangerousAction(ctx context.Context, action, selector, value, reasoning string) (bool, string, error) {
	prompt := fmt.Sprintf(`Ты - система безопасности для AI-агента, который управляет браузером.

Агент планирует выполнить действие:
- Действие: %s
- Селектор: %s
- Значение: %s
- Обоснование: %s

Определи, является ли это действие потенциально опасным. Опасными считаются действия, которые могут:
- Привести к финансовым операциям (покупка, оплата, перевод)
- Удалить данные (удаление файлов, писем, аккаунтов)
- Отправить конфиденциальную информацию
- Подтвердить необратимые действия

Ответь в формате JSON:
{
  "is_dangerous": true/false,
  "reason": "краткое объяснение почему действие опасно или безопасно",
  "message": "сообщение для пользователя с деталями действия"
}`, action, selector, value, reasoning)

	systemMsg := "Ты - система безопасности. Анализируй действия AI-агента и определяй потенциальные риски."

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
		Temperature: 0.3,
	})

	if err != nil {
		return false, "", fmt.Errorf("ошибка запроса к OpenAI для проверки безопасности: %w", err)
	}

	if len(resp.Choices) == 0 {
		return false, "", fmt.Errorf("пустой ответ от OpenAI")
	}

	responseText := resp.Choices[0].Message.Content

	var result SecurityCheckResult
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return false, "", fmt.Errorf("ошибка парсинга ответа безопасности: %w", err)
	}

	if c.logger != nil {
		sanitizedPrompt := c.sanitizer.Sanitize(prompt)
		sanitizedResponse := c.sanitizer.Sanitize(responseText)
		_ = c.logger.LogLLMRequest(ctx, nil, nil, "security_check", sanitizedPrompt, sanitizedResponse, c.model, resp.Usage.TotalTokens)
	}

	return result.IsDangerous, result.Message, nil
}
