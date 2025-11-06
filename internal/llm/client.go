package llm

import (
	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	model  string
	logger Logger
}

func NewClient(apiKey, model string, logger Logger) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
		model:  model,
		logger: logger,
	}
}
