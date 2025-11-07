package llm

import (
	"context"

	"aiAgent/internal/sanitizer"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client      *openai.Client
	model       string
	logger      Logger
	sanitizer   *sanitizer.DataSanitizer
	rateLimiter *RateLimiter
}

func NewClient(apiKey, model string, logger Logger) *Client {
	return NewClientWithRateLimit(apiKey, model, logger, 60, 90000)
}

func NewClientWithRateLimit(apiKey, model string, logger Logger, requestsPerMinute, tokensPerHour int) *Client {
	client := &Client{
		client:      openai.NewClient(apiKey),
		model:       model,
		logger:      logger,
		rateLimiter: NewRateLimiter(requestsPerMinute, tokensPerHour),
	}

	client.sanitizer = sanitizer.NewWithAI(client)
	return client
}

// createChatCompletionWithRateLimit выполняет запрос с проверкой rate limit
func (c *Client) createChatCompletionWithRateLimit(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	// Проверка лимита запросов
	if err := c.rateLimiter.AllowRequest(ctx); err != nil {
		return openai.ChatCompletionResponse{}, err
	}

	// Оценка количества токенов (грубая оценка: ~4 символа на токен)
	estimatedTokens := 0
	for _, msg := range req.Messages {
		estimatedTokens += len(msg.Content) / 4
	}
	estimatedTokens += req.MaxTokens // добавляем токены ответа

	// Проверка лимита токенов
	if err := c.rateLimiter.AllowTokens(ctx, estimatedTokens); err != nil {
		return openai.ChatCompletionResponse{}, err
	}

	// Выполняем запрос
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return resp, err
	}

	// Корректируем использованные токены (теперь знаем точное значение)
	if resp.Usage.TotalTokens > estimatedTokens {
		c.rateLimiter.ConsumeTokens(resp.Usage.TotalTokens - estimatedTokens)
	}

	return resp, nil
}
