package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

func (c *Client) CheckSensitiveData(ctx context.Context, text string) (bool, error) {
	prompt := fmt.Sprintf(`Ты - система безопасности для маскирования персональных данных.

Проанализируй следующий текст и определи, содержит ли он персональные или чувствительные данные:
- Пароли, токены, API ключи
- Номера телефонов, email адреса
- Адреса проживания
- Номера банковских карт, CVV
- Другие персональные данные

Текст для анализа:
%s

Ответь в формате JSON:
{
  "is_sensitive": true/false,
  "reason": "краткое объяснение"
}`, text)

	systemMsg := "Ты - система безопасности. Определяй наличие персональных данных в тексте."

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
		Temperature: 0.1,
		MaxTokens:   100,
	})

	if err != nil {
		return false, fmt.Errorf("ошибка запроса к OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return false, fmt.Errorf("пустой ответ от OpenAI")
	}

	responseText := resp.Choices[0].Message.Content

	var result struct {
		IsSensitive bool   `json:"is_sensitive"`
		Reason      string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return false, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return result.IsSensitive, nil
}
