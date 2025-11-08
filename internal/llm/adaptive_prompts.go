package llm

import "strings"

type TaskCategory string

const (
	CategoryNavigation TaskCategory = "navigation"
	CategoryForm       TaskCategory = "form"
	CategoryExtraction TaskCategory = "extraction"
	CategoryPurchase   TaskCategory = "purchase"
	CategoryGeneral    TaskCategory = "general"
)

func DetectTaskCategory(task string) TaskCategory {
	taskLower := strings.ToLower(task)

	navigationKeywords := []string{"открой", "перейди", "найди страницу", "navigate", "go to", "open", "visit"}
	for _, kw := range navigationKeywords {
		if strings.Contains(taskLower, kw) {
			return CategoryNavigation
		}
	}

	formKeywords := []string{"заполни", "введи", "отправь форму", "зарегистрируйся", "войди", "fill", "submit", "login", "register"}
	for _, kw := range formKeywords {
		if strings.Contains(taskLower, kw) {
			return CategoryForm
		}
	}

	extractionKeywords := []string{"найди", "извлеки", "получи информацию", "скопируй", "прочитай", "find", "extract", "get", "read"}
	for _, kw := range extractionKeywords {
		if strings.Contains(taskLower, kw) {
			return CategoryExtraction
		}
	}

	purchaseKeywords := []string{"купи", "оплати", "закажи", "добавь в корзину", "buy", "purchase", "order", "checkout", "cart"}
	for _, kw := range purchaseKeywords {
		if strings.Contains(taskLower, kw) {
			return CategoryPurchase
		}
	}

	return CategoryGeneral
}

// GetSystemPromptForCategory возвращает МИНИМАЛЬНЫЙ system prompt.
// С внедрением Reasoning Layer детальные инструкции больше не нужны -
// агент сам рассуждает о том как действовать.
func GetSystemPromptForCategory(category TaskCategory) string {
	// Единый минимальный промпт для всех категорий
	// Reasoning layer теперь отвечает за анализ и стратегию
	return `Ты автономный AI-агент для управления браузером.

Твоя задача - планировать конкретные действия на основе:
- Текущей задачи пользователя
- Контекста страницы (доступные элементы)
- Стратегии из reasoning (если доступна)

Доступные действия:
- navigate(url) - переход по URL
- click(selector) - клик по элементу
- type(selector, value) - ввод текста
- extract_info(selector) - извлечение информации
- ask_user(question) - запрос у пользователя
- complete() - задача выполнена

Используй tool calling для выбора действия.
Отвечай на русском языке.`
}

// GetFewShotExamplesForCategory больше не используется с Reasoning Layer.
// Few-shot примеры теперь не нужны - агент учится через reasoning и память.
// Сохранено для обратной совместимости, но возвращает пустую строку.
func GetFewShotExamplesForCategory(category TaskCategory) string {
	// С Reasoning Layer агент сам вырабатывает стратегию
	// Few-shot примеры создают "заготовки действий" - это против требований ТЗ
	return ""
}

// GetTaskSpecificGuidance больше не используется с Reasoning Layer.
// Специфичные руководства теперь не нужны - reasoning определяет что важно.
// Сохранено для обратной совместимости, но возвращает пустую строку.
func GetTaskSpecificGuidance(category TaskCategory) string {
	// С Reasoning Layer агент сам определяет ключевые соображения
	// Task-specific guidance = микроменеджмент, против philosophy нового подхода
	return ""
}
