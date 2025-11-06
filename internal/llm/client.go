package llm

import (
	"aiAgent/internal/sanitizer"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client    *openai.Client
	model     string
	logger    Logger
	sanitizer *sanitizer.DataSanitizer
}

func NewClient(apiKey, model string, logger Logger) *Client {
	client := &Client{
		client: openai.NewClient(apiKey),
		model:  model,
		logger: logger,
	}

	client.sanitizer = sanitizer.NewWithAI(client)
	return client
}
