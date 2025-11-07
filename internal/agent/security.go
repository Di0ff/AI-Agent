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

// DangerousPattern определяет паттерн опасного действия
type DangerousPattern struct {
	Keywords    []string // ключевые слова для проверки
	Description string   // описание опасности
	Severity    string   // уровень опасности: "high", "medium", "low"
}

// Паттерны опасных действий
var dangerousPatterns = []DangerousPattern{
	// Финансовые операции
	{
		Keywords:    []string{"pay", "payment", "купить", "оплат", "card", "карт", "checkout", "purchase", "buy", "transaction"},
		Description: "Финансовая операция или платеж",
		Severity:    "high",
	},
	// Удаление данных
	{
		Keywords:    []string{"delete", "remove", "удал", "очист", "trash", "erase", "destroy"},
		Description: "Удаление данных или контента",
		Severity:    "high",
	},
	// Изменение критичных настроек
	{
		Keywords:    []string{"password", "пароль", "security", "безопасност", "account", "аккаунт", "settings", "настройк"},
		Description: "Изменение критичных настроек аккаунта",
		Severity:    "high",
	},
	// Отправка данных
	{
		Keywords:    []string{"submit", "send", "отправ", "transfer", "перевод", "share", "публик", "publish"},
		Description: "Отправка или публикация данных",
		Severity:    "medium",
	},
	// Подтверждение действий
	{
		Keywords:    []string{"confirm", "подтверд", "accept", "принять", "agree", "соглас"},
		Description: "Подтверждение необратимого действия",
		Severity:    "medium",
	},
}

func NewSecurityChecker(llmClient llm.LLMClient) *SecurityChecker {
	return &SecurityChecker{
		llmClient: llmClient,
	}
}

// checkRuleBasedDanger выполняет быструю проверку по правилам без LLM
func (s *SecurityChecker) checkRuleBasedDanger(action, selector, value, reasoning string) (bool, string) {
	// Объединяем все данные для проверки
	combined := strings.ToLower(action + " " + selector + " " + value + " " + reasoning)

	for _, pattern := range dangerousPatterns {
		for _, keyword := range pattern.Keywords {
			if strings.Contains(combined, keyword) {
				// Дополнительная проверка на контекст - избегаем false positives
				if pattern.Severity == "high" {
					return true, fmt.Sprintf("Обнаружен паттерн опасного действия: %s (ключевое слово: '%s')", pattern.Description, keyword)
				}
			}
		}
	}

	// Проверка на опасные селекторы
	if s.isDangerousSelector(selector) {
		return true, "Обнаружен селектор критичного элемента (кнопка удаления, подтверждения платежа и т.д.)"
	}

	return false, ""
}

// isDangerousSelector проверяет, является ли селектор потенциально опасным
func (s *SecurityChecker) isDangerousSelector(selector string) bool {
	selectorLower := strings.ToLower(selector)

	dangerousSelectors := []string{
		"delete", "удал", "remove", "trash",
		"confirm", "подтверд", "accept",
		"payment", "оплат", "checkout", "buy",
	}

	for _, dangerous := range dangerousSelectors {
		if strings.Contains(selectorLower, dangerous) {
			return true
		}
	}

	// Проверка на submit кнопки в формах с определенными паттернами
	if strings.Contains(selectorLower, "submit") && (strings.Contains(selectorLower, "payment") || strings.Contains(selectorLower, "delete")) {
		return true
	}

	return false
}

func (s *SecurityChecker) IsDangerousAction(ctx context.Context, action, selector, value, reasoning string) (bool, string, error) {
	// Шаг 0: Проверка домена для navigate действий
	if action == "navigate" && value != "" {
		domainSec := CheckDomainSecurity(value)

		if domainSec.Level == DomainBlocked {
			return true, fmt.Sprintf("ЗАБЛОКИРОВАННЫЙ ДОМЕН: %s. %s", domainSec.Description, domainSec.Reason), nil
		}

		if domainSec.Level == DomainCritical {
			return true, fmt.Sprintf("КРИТИЧНЫЙ ДОМЕН: %s. %s", domainSec.Description, domainSec.Reason), nil
		}
	}

	// Шаг 1: Быстрая проверка по правилам (без LLM)
	isDangerous, ruleMessage := s.checkRuleBasedDanger(action, selector, value, reasoning)

	// Если правила обнаружили опасность и есть LLM - дополнительно проверяем через LLM
	if isDangerous && s.llmClient != nil {
		// Двойная проверка через LLM для уточнения
		llmDangerous, llmMessage, err := s.llmClient.CheckDangerousAction(ctx, action, selector, value, reasoning)
		if err != nil {
			// Если LLM недоступен, используем результат правил
			return true, ruleMessage, nil
		}

		// Если LLM подтвердил опасность, используем его сообщение (оно более детальное)
		if llmDangerous {
			return true, llmMessage, nil
		}

		// Если LLM не подтвердил, но правила сработали - используем правила (более строгий подход)
		return true, ruleMessage, nil
	}

	// Если правила не обнаружили опасность, возвращаем безопасно
	if !isDangerous {
		return false, "", nil
	}

	// Если правила обнаружили опасность, но нет LLM клиента
	if s.llmClient == nil {
		return true, ruleMessage, nil
	}

	return false, "", nil
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
