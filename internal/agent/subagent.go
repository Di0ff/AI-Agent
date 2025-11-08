package agent

import (
	"aiAgent/internal/llm"
	"context"
	"fmt"
)

type LLMClient = llm.LLMClient

type TaskType string

const (
	TaskTypeEmailSpam    TaskType = "email_spam"
	TaskTypeFoodDelivery TaskType = "food_delivery"
	TaskTypeJobSearch    TaskType = "job_search"
	TaskTypeGeneral      TaskType = "general"
)

type SpecializedAgent interface {
	CanHandle(ctx context.Context, task string, pageContext string) (confidence float64, err error)
	Execute(ctx context.Context, task string, maxSteps int) error
	GetExpertise() []string
	GetType() TaskType
	GetDescription() string
}

type AgentRouter struct {
	agents        map[TaskType]SpecializedAgent
	defaultAgent  SpecializedAgent
	llmClient     LLMClient
	confidenceMin float64
}

func NewAgentRouter(llmClient LLMClient, confidenceMin float64) *AgentRouter {
	if confidenceMin == 0 {
		confidenceMin = 0.7
	}

	return &AgentRouter{
		agents:        make(map[TaskType]SpecializedAgent),
		llmClient:     llmClient,
		confidenceMin: confidenceMin,
	}
}

func (r *AgentRouter) RegisterAgent(agent SpecializedAgent) {
	r.agents[agent.GetType()] = agent
}

func (r *AgentRouter) SetDefaultAgent(agent SpecializedAgent) {
	r.defaultAgent = agent
}

func (r *AgentRouter) RouteTask(ctx context.Context, task string, pageContext string) (SpecializedAgent, error) {
	bestAgent := r.defaultAgent
	bestConfidence := 0.0

	for _, agent := range r.agents {
		confidence, err := agent.CanHandle(ctx, task, pageContext)
		if err != nil {
			continue
		}

		if confidence > bestConfidence {
			bestConfidence = confidence
			bestAgent = agent
		}
	}

	if bestConfidence < r.confidenceMin && r.defaultAgent != nil {
		return r.defaultAgent, nil
	}

	if bestAgent == nil {
		return nil, fmt.Errorf("no suitable agent found for task: %s", task)
	}

	return bestAgent, nil
}

func (r *AgentRouter) ExecuteWithRouting(ctx context.Context, task string, pageContext string, maxSteps int) error {
	agent, err := r.RouteTask(ctx, task, pageContext)
	if err != nil {
		return fmt.Errorf("failed to route task: %w", err)
	}

	return agent.Execute(ctx, task, maxSteps)
}

func (r *AgentRouter) ListAgents() []SpecializedAgent {
	agents := make([]SpecializedAgent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}
