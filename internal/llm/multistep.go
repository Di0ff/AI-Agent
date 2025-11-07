package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type MultiStepPlan struct {
	Steps            []StepPlan `json:"steps"`
	OverallStrategy  string     `json:"overall_strategy"`
	FallbackStrategy string     `json:"fallback_strategy"`
	EstimatedSteps   int        `json:"estimated_steps"`
}

func (c *Client) PlanMultiStep(ctx context.Context, task string, pageContext string, maxSteps int, taskID *uint, stepID *uint) (*MultiStepPlan, error) {
	if maxSteps == 0 {
		maxSteps = 5
	}

	category := DetectTaskCategory(task)
	systemMsg := GetSystemPromptForCategory(category)
	fewShot := GetFewShotExamplesForCategory(category)
	guidance := GetTaskSpecificGuidance(category)

	prompt := fmt.Sprintf(`You are an AI agent controlling a web browser. Plan a sequence of actions to accomplish this task.

Task: %s

Current page context:
%s

%s

%s

Plan the next %d steps to accomplish this task. Think strategically about the optimal sequence.

Available actions:
- navigate: go to a URL
- click: click an element
- type: type text into an element
- extract_info: extract information from the page
- ask_user: ask the user for information
- complete: task is finished

Respond in JSON format:
{
  "steps": [
    {
      "action": "action_name",
      "selector": "css_selector (if applicable)",
      "value": "value (if applicable)",
      "reasoning": "why this step"
    }
  ],
  "overall_strategy": "high-level strategy description",
  "fallback_strategy": "what to do if a step fails",
  "estimated_steps": number
}`, task, pageContext, fewShot, guidance, maxSteps)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: systemMsg + "\n\nYou are an expert at planning multi-step web automation tasks. Think strategically and plan ahead.",
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
		return nil, fmt.Errorf("failed to plan multi-step: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	content := resp.Choices[0].Message.Content
	var result MultiStepPlan
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse multi-step plan: %w", err)
	}

	if c.logger != nil {
		c.logger.LogLLMRequest(ctx, taskID, stepID, "system", prompt, content, c.model, resp.Usage.TotalTokens)
	}

	return &result, nil
}

func (c *Client) Replan(ctx context.Context, task string, pageContext string, originalPlan *MultiStepPlan, failedStep *StepPlan, errorMessage string, maxSteps int, taskID *uint, stepID *uint) (*MultiStepPlan, error) {
	if maxSteps == 0 {
		maxSteps = 5
	}

	stepsJSON, _ := json.MarshalIndent(originalPlan.Steps, "", "  ")

	prompt := fmt.Sprintf(`You are an AI agent controlling a web browser. The original plan failed, you need to replan.

Task: %s

Current page context:
%s

Original plan:
%s

Failed step:
Action: %s
Selector: %s
Value: %s
Reasoning: %s

Error: %s

Create a new plan that works around this failure. Consider alternative approaches.

Plan the next %d steps. Respond in JSON format:
{
  "steps": [
    {
      "action": "action_name",
      "selector": "css_selector (if applicable)",
      "value": "value (if applicable)",
      "reasoning": "why this step"
    }
  ],
  "overall_strategy": "new strategy description",
  "fallback_strategy": "what to do if this fails",
  "estimated_steps": number
}`, task, pageContext, string(stepsJSON), failedStep.Action, failedStep.Selector, failedStep.Value, failedStep.Reasoning, errorMessage, maxSteps)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "You are an expert at recovering from failures and finding alternative approaches to web automation tasks.",
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
		return nil, fmt.Errorf("failed to replan: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	content := resp.Choices[0].Message.Content
	var result MultiStepPlan
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse replan: %w", err)
	}

	if c.logger != nil {
		c.logger.LogLLMRequest(ctx, taskID, stepID, "system", prompt, content, c.model, resp.Usage.TotalTokens)
	}

	return &result, nil
}
