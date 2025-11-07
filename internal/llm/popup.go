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
	prompt := fmt.Sprintf(`Analyze the page elements and determine if there is a popup, modal, or overlay that should be closed.

Elements data:
%s

Determine:
1. Is there a popup/modal/overlay present?
2. If yes, what is the CSS selector of the close button?
3. Brief description of the popup

Respond in JSON format:
{
  "has_popup": true/false,
  "close_selector": "CSS selector",
  "popup_description": "brief description",
  "reasoning": "your analysis"
}`, elements)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "You are an expert at analyzing web page structure and identifying popups and their close buttons.",
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
