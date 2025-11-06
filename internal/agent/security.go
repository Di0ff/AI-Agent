package agent

import (
	"context"
	"fmt"
	"strings"

	"aiAgent/internal/llm"
)

type SecurityChecker struct {
	llmClient llm.LLMClient
}

func NewSecurityChecker(llmClient llm.LLMClient) *SecurityChecker {
	return &SecurityChecker{
		llmClient: llmClient,
	}
}

func (s *SecurityChecker) IsDangerousAction(ctx context.Context, action, selector, value, reasoning string) (bool, string, error) {
	if s.llmClient == nil {
		return false, "", fmt.Errorf("LLM клиент не настроен для проверки безопасности")
	}

	return s.llmClient.CheckDangerousAction(ctx, action, selector, value, reasoning)
}

func (s *SecurityChecker) GetConfirmationMessage(action string, selector string, value string, reasoning string, llmMessage string) string {
	actionLower := action

	baseMsg := ""
	if strings.Contains(actionLower, "click") {
		baseMsg = fmt.Sprintf("ВНИМАНИЕ: Агент хочет выполнить потенциально опасное действие - клик по элементу.\nСелектор: %s\nОбоснование: %s", selector, reasoning)
	} else if strings.Contains(actionLower, "type") {
		baseMsg = fmt.Sprintf("ВНИМАНИЕ: Агент хочет ввести данные в поле.\nСелектор: %s\nЗначение: %s\nОбоснование: %s", selector, value, reasoning)
	} else if strings.Contains(actionLower, "navigate") {
		baseMsg = fmt.Sprintf("ВНИМАНИЕ: Агент хочет перейти на страницу.\nURL: %s\nОбоснование: %s", value, reasoning)
	} else {
		baseMsg = fmt.Sprintf("ВНИМАНИЕ: Агент хочет выполнить потенциально опасное действие.\nДействие: %s\nОбоснование: %s", action, reasoning)
	}

	if llmMessage != "" {
		baseMsg += "\n\n" + llmMessage
	}

	return baseMsg + "\n\nПродолжить? (yes/no): "
}
