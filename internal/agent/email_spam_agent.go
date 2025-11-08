package agent

import (
	"context"
	"strings"
)

// EmailSpamAgent - специализированный агент для работы с почтой и удаления спама
// Ключевые слова: "почта", "письма", "спам", "email", "inbox", "яндекс почта", "gmail"
type EmailSpamAgent struct {
	baseAgent *Agent
	keywords  []string
}

// NewEmailSpamAgent создает новый EmailSpamAgent
func NewEmailSpamAgent(baseAgent *Agent) *EmailSpamAgent {
	return &EmailSpamAgent{
		baseAgent: baseAgent,
		keywords: []string{
			// Русские ключевые слова
			"почта", "почтовый", "письма", "письмо", "спам", "inbox",
			"яндекс почта", "яндекс.почта", "gmail", "почтовый ящик",
			// Английские ключевые слова
			"email", "mail", "message", "messages", "spam",
		},
	}
}

// CanHandle проверяет, может ли EmailSpamAgent обработать данную задачу
func (a *EmailSpamAgent) CanHandle(ctx context.Context, task string, pageContext string) (float64, error) {
	taskLower := strings.ToLower(task)
	contextLower := strings.ToLower(pageContext)

	// Проверяем ключевые слова в задаче
	taskScore := 0.0
	for _, keyword := range a.keywords {
		if strings.Contains(taskLower, keyword) {
			taskScore = 0.9
			break
		}
	}

	// Проверяем контекст страницы (почтовые сервисы)
	contextScore := 0.0
	mailIndicators := []string{
		"mail.yandex", "gmail.com", "mail.google", "inbox",
		"письма", "входящие", "спам", "email",
	}
	for _, indicator := range mailIndicators {
		if strings.Contains(contextLower, indicator) {
			contextScore = 0.3
			break
		}
	}

	return taskScore + contextScore, nil
}

// Execute выполняет задачу через базового агента
func (a *EmailSpamAgent) Execute(ctx context.Context, task string, maxSteps int) error {
	return a.baseAgent.executeTaskString(ctx, task, maxSteps)
}

// GetExpertise возвращает список экспертиз агента
func (a *EmailSpamAgent) GetExpertise() []string {
	return []string{
		"email_management",
		"spam_detection",
		"mail_organization",
		"yandex_mail",
		"gmail",
	}
}

// GetType возвращает тип задачи
func (a *EmailSpamAgent) GetType() TaskType {
	return TaskTypeEmailSpam
}

// GetDescription возвращает описание агента
func (a *EmailSpamAgent) GetDescription() string {
	return "Специализированный агент для работы с почтой: чтение писем, определение и удаление спама"
}
