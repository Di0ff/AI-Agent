// Package llm - reasoning module для реализации ReAct pattern (Reasoning + Acting).
// Отделяет фазу рассуждения от фазы планирования действий.
package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// ReasoningStep представляет один шаг явного рассуждения агента.
// Используется для реализации ReAct pattern - агент сначала рассуждает, потом действует.
type ReasoningStep struct {
	// Observation - что агент видит/понимает в текущей ситуации
	Observation string `json:"observation"`

	// Analysis - анализ текущей ситуации и контекста
	Analysis string `json:"analysis"`

	// Strategy - выбранная стратегия действий
	Strategy string `json:"strategy"`

	// Confidence - уровень уверенности в выбранной стратегии (0.0 - 1.0)
	Confidence float64 `json:"confidence"`

	// Alternatives - рассмотренные альтернативные подходы
	Alternatives []string `json:"alternatives,omitempty"`

	// Uncertainties - что агент не знает или не уверен
	Uncertainties []string `json:"uncertainties,omitempty"`

	// RequiresUserInput - нужна ли дополнительная информация от пользователя
	RequiresUserInput bool `json:"requires_user_input"`

	// ReasonForUserInput - почему нужен ввод пользователя
	ReasonForUserInput string `json:"reason_for_user_input,omitempty"`
}

// ReasoningHistory хранит историю рассуждений агента.
// Используется для передачи контекста между шагами.
type ReasoningHistory struct {
	Steps []ReasoningStep `json:"steps"`
}

// AddStep добавляет новый шаг рассуждения в историю.
func (rh *ReasoningHistory) AddStep(step ReasoningStep) {
	rh.Steps = append(rh.Steps, step)
}

// GetLastStep возвращает последний шаг рассуждения или nil если история пуста.
func (rh *ReasoningHistory) GetLastStep() *ReasoningStep {
	if len(rh.Steps) == 0 {
		return nil
	}
	return &rh.Steps[len(rh.Steps)-1]
}

// Clear очищает историю рассуждений.
func (rh *ReasoningHistory) Clear() {
	rh.Steps = nil
}

// ToJSON сериализует историю в JSON для передачи в контексте LLM.
func (rh *ReasoningHistory) ToJSON() string {
	if len(rh.Steps) == 0 {
		return "[]"
	}
	data, err := json.MarshalIndent(rh.Steps, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(data)
}

// Reason выполняет фазу explicit reasoning перед планированием действия.
// Это ключевой метод ReAct pattern - агент сначала явно рассуждает о ситуации,
// а только потом планирует конкретное действие.
//
// Параметры:
//   - ctx: контекст выполнения
//   - task: текущая задача пользователя
//   - pageContext: контекст страницы (snapshot элементов)
//   - history: история предыдущих рассуждений (может быть nil)
//   - taskID, stepID: идентификаторы для логирования
//
// Возвращает:
//   - ReasoningStep с полями observation, analysis, strategy, confidence
//   - error в случае ошибки LLM запроса
func (c *Client) Reason(ctx context.Context, task string, pageContext string, history *ReasoningHistory, taskID *uint, stepID *uint) (*ReasoningStep, error) {
	// Минимальный system prompt - только роль и формат ответа
	systemPrompt := `Ты автономный AI-агент для управления браузером.

Твоя задача - проанализировать текущую ситуацию и выработать стратегию действий.

НЕ планируй конкретные действия - только рассуждай о том:
- Что ты видишь (observation)
- Как это анализируешь (analysis)
- Какую общую стратегию выберешь (strategy)
- Насколько уверен в этой стратегии (confidence: 0.0-1.0)
- Какие альтернативы рассматривал (alternatives)
- Что не понятно или вызывает неуверенность (uncertainties)
- Нужна ли дополнительная информация от пользователя (requires_user_input)

Отвечай ТОЛЬКО в формате JSON со следующей структурой:
{
  "observation": "что ты видишь...",
  "analysis": "твой анализ ситуации...",
  "strategy": "общая стратегия без конкретных действий...",
  "confidence": 0.8,
  "alternatives": ["альтернатива 1", "альтернатива 2"],
  "uncertainties": ["неясность 1"],
  "requires_user_input": false,
  "reason_for_user_input": ""
}`

	// Формируем user prompt с задачей и контекстом
	userPrompt := fmt.Sprintf(`Текущая задача: %s

Контекст страницы:
%s`, task, pageContext)

	// Добавляем историю рассуждений если есть
	if history != nil && len(history.Steps) > 0 {
		userPrompt += fmt.Sprintf(`

История предыдущих рассуждений:
%s`, history.ToJSON())
	}

	// Делаем запрос к LLM в JSON mode для получения structured output
	resp, err := c.createChatCompletionWithRateLimit(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Temperature: 0.7, // Немного creativity для reasoning
	})

	if err != nil {
		if c.logger != nil {
			fullPrompt := formatPrompt(systemPrompt, userPrompt)
			sanitizedPrompt := c.sanitizer.Sanitize(fullPrompt)
			sanitizedError := c.sanitizer.Sanitize(err.Error())
			_ = c.logger.LogLLMRequest(ctx, taskID, stepID, "reasoning_error", sanitizedPrompt, sanitizedError, c.model, 0)
		}
		return nil, fmt.Errorf("ошибка reasoning запроса к OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("пустой ответ reasoning от OpenAI")
	}

	// Парсим JSON ответ в ReasoningStep
	responseText := resp.Choices[0].Message.Content
	var reasoning ReasoningStep
	if err := json.Unmarshal([]byte(responseText), &reasoning); err != nil {
		if c.logger != nil {
			fullPrompt := formatPrompt(systemPrompt, userPrompt)
			sanitizedPrompt := c.sanitizer.Sanitize(fullPrompt)
			sanitizedResponse := c.sanitizer.Sanitize(responseText)
			_ = c.logger.LogLLMRequest(ctx, taskID, stepID, "reasoning_parse_error", sanitizedPrompt, sanitizedResponse, c.model, resp.Usage.TotalTokens)
		}
		return nil, fmt.Errorf("ошибка парсинга reasoning JSON: %w", err)
	}

	// Логируем успешный reasoning
	if c.logger != nil {
		fullPrompt := formatPrompt(systemPrompt, userPrompt)
		sanitizedPrompt := c.sanitizer.Sanitize(fullPrompt)
		sanitizedResponse := c.sanitizer.Sanitize(responseText)
		_ = c.logger.LogLLMRequest(ctx, taskID, stepID, "reasoning", sanitizedPrompt, sanitizedResponse, c.model, resp.Usage.TotalTokens)
	}

	return &reasoning, nil
}

// ReasonWithContext выполняет reasoning с дополнительным контекстом из Memory.
// Используется когда есть relevant patterns из предыдущего опыта.
func (c *Client) ReasonWithContext(ctx context.Context, task string, pageContext string, history *ReasoningHistory, memoryContext string, taskID *uint, stepID *uint) (*ReasoningStep, error) {
	// Аналогично Reason(), но добавляем memory context в prompt
	systemPrompt := `Ты автономный AI-агент для управления браузером.

Твоя задача - проанализировать текущую ситуацию и выработать стратегию действий.

НЕ планируй конкретные действия - только рассуждай о том:
- Что ты видишь (observation)
- Как это анализируешь (analysis)
- Какую общую стратегию выберешь (strategy)
- Насколько уверен в этой стратегии (confidence: 0.0-1.0)
- Какие альтернативы рассматривал (alternatives)
- Что не понятно или вызывает неуверенность (uncertainties)
- Нужна ли дополнительная информация от пользователя (requires_user_input)

Учитывай опыт из похожих ситуаций, но адаптируй стратегию под текущий контекст.

Отвечай ТОЛЬКО в формате JSON со следующей структурой:
{
  "observation": "что ты видишь...",
  "analysis": "твой анализ ситуации...",
  "strategy": "общая стратегия без конкретных действий...",
  "confidence": 0.8,
  "alternatives": ["альтернатива 1", "альтернатива 2"],
  "uncertainties": ["неясность 1"],
  "requires_user_input": false,
  "reason_for_user_input": ""
}`

	userPrompt := fmt.Sprintf(`Текущая задача: %s

Контекст страницы:
%s

Релевантный опыт из памяти:
%s`, task, pageContext, memoryContext)

	if history != nil && len(history.Steps) > 0 {
		userPrompt += fmt.Sprintf(`

История предыдущих рассуждений:
%s`, history.ToJSON())
	}

	resp, err := c.createChatCompletionWithRateLimit(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Temperature: 0.7,
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка reasoning запроса к OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("пустой ответ reasoning от OpenAI")
	}

	responseText := resp.Choices[0].Message.Content
	var reasoning ReasoningStep
	if err := json.Unmarshal([]byte(responseText), &reasoning); err != nil {
		return nil, fmt.Errorf("ошибка парсинга reasoning JSON: %w", err)
	}

	if c.logger != nil {
		fullPrompt := formatPrompt(systemPrompt, userPrompt)
		sanitizedPrompt := c.sanitizer.Sanitize(fullPrompt)
		sanitizedResponse := c.sanitizer.Sanitize(responseText)
		_ = c.logger.LogLLMRequest(ctx, taskID, stepID, "reasoning_with_context", sanitizedPrompt, sanitizedResponse, c.model, resp.Usage.TotalTokens)
	}

	return &reasoning, nil
}
