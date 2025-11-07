package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type PopupInfo struct {
	HasPopup         bool   `json:"has_popup"`
	CloseSelector    string `json:"close_selector"`
	PopupDescription string `json:"popup_description"`
	Reasoning        string `json:"reasoning"`
}

func (c *Client) AnalyzePopup(ctx context.Context, elements string) (*PopupInfo, error) {
	prompt := fmt.Sprintf(`Проанализируй элементы страницы и определи есть ли всплывающее окно, модальное окно или оверлей которые нужно закрыть.

Данные элементов:
%s

Определи:
1. Есть ли всплывающее окно/модальное окно/оверлей?
2. Если да, какой CSS селектор кнопки закрытия?
3. Краткое описание всплывающего окна

Отвечай в формате JSON:
{
  "has_popup": true/false,
  "close_selector": "CSS селектор",
  "popup_description": "краткое описание",
  "reasoning": "твой анализ"
}`, elements)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "Ты эксперт в анализе структуры веб-страниц и определении всплывающих окон и их кнопок закрытия. ВСЕГДА отвечай на русском языке.",
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
		return nil, fmt.Errorf("failed to analyze popup: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	content := resp.Choices[0].Message.Content
	var result PopupInfo
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse popup analysis: %w", err)
	}

	if c.logger != nil {
		c.logger.LogLLMRequest(ctx, nil, nil, "system", prompt, content, c.model, resp.Usage.TotalTokens)
	}

	return &result, nil
}
