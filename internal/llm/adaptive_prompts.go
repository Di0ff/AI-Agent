package llm

import "strings"

type TaskCategory string

const (
	CategoryNavigation TaskCategory = "navigation"
	CategoryForm       TaskCategory = "form"
	CategoryExtraction TaskCategory = "extraction"
	CategoryPurchase   TaskCategory = "purchase"
	CategoryEmail      TaskCategory = "email"
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

	// Проверяем почтовые задачи ПЕРЕД extraction (так как "прочитай" может быть и про почту)
	emailKeywords := []string{"почта", "письма", "письмо", "спам", "email", "mail", "inbox", "яндекс почта", "gmail"}
	for _, kw := range emailKeywords {
		if strings.Contains(taskLower, kw) {
			return CategoryEmail
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
	// Для почтовых задач добавляем специализированные инструкции
	if category == CategoryEmail {
		return `Ты автономный AI-агент для управления браузером, специализирующийся на работе с почтой.

Твоя задача - планировать конкретные действия на основе:
- Текущей задачи пользователя (работа с почтой: чтение, удаление спама)
- Контекста страницы (доступные элементы)
- Стратегии из reasoning (если доступна)

ВАЖНО: ДЕЙСТВУЙ, а не спрашивай! Используй свои знания для определения спама.

Для работы с почтой (Яндекс.Почта, Gmail):
1. Если не на странице почты - перейди на mail.yandex.ru или gmail.com
2. Найди папку "Входящие" или список писем (обычно уже открыта)
3. КРИТИЧНО: Для извлечения информации о письмах:
   - НЕ используй селекторы аватаров (a[data-testid="messages-list_message-avatar"]) - это только иконки!
   - СТРАТЕГИЯ 1: Попробуй кликнуть на первое письмо в списке, чтобы открыть его и увидеть тему, отправителя, содержание
   - СТРАТЕГИЯ 2: Если письма открываются в правой панели, используй extract_info на элементах письма:
     * Тема письма: обычно в заголовке или .mail-Message-Head-Theme
     * Отправитель: обычно .mail-Message-Head-From или [data-testid*="sender"]
     * Содержание: обычно .mail-Message-Body или .mail-Message-Content
   - СТРАТЕГИЯ 3: Если видишь список писем в левой панели, ищи элементы с текстом (тема, отправитель видны в списке)
   - СТРАТЕГИЯ 4: Прокрути страницу вниз, чтобы увидеть больше писем
4. Прочитай последние 10 писем (открой каждое письмо, извлеки тему, отправителя, краткое содержание)
5. Определи спам-письма на основе общих признаков:
   - Рекламные рассылки (promo, акции, скидки в теме)
   - Подозрительные отправители (noreply, no-reply, случайные домены)
   - Фишинг (требования паролей, срочные действия)
   - Массовые рассылки от неизвестных отправителей
6. Выбери спам-письма (используй чекбоксы или кликни на них) и удали их (используй кнопку "Удалить")
7. Предоставь отчет: сколько спама удалено, какие важные письма остались

КРИТИЧНО: НЕ спрашивай пользователя о критериях спама - используй свои знания и действуй!
Используй ask_user ТОЛЬКО в крайних случаях, когда действительно невозможно определить.

Доступные действия:
- navigate(url) - переход по URL
- click(selector) - клик по элементу
- type(selector, value) - ввод текста
- extract_info(selector) - извлечение информации со страницы (ИСПОЛЬЗУЙ ПРАВИЛЬНЫЕ СЕЛЕКТОРЫ!)
- ask_user(question) - запрос у пользователя (ИСПОЛЬЗУЙ МИНИМАЛЬНО!)
- complete() - задача выполнена

Используй tool calling для выбора действия.
Отвечай на русском языке.`
	}

	// Единый минимальный промпт для остальных категорий
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
