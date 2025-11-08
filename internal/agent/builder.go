package agent

import "strings"

func (a *Agent) buildLimitedContext(elements []PageElement) string {
	var parts []string

	critical := []PageElement{}
	high := []PageElement{}
	medium := []PageElement{}
	low := []PageElement{}

	for _, el := range elements {
		switch el.Priority {
		case PriorityCritical:
			critical = append(critical, el)
		case PriorityHigh:
			high = append(high, el)
		case PriorityMedium:
			medium = append(medium, el)
		default:
			low = append(low, el)
		}
	}

	maxElements := a.maxTokens / 50

	if len(critical) > 0 {
		for i := 0; i < len(critical) && i < maxElements; i++ {
			parts = append(parts, formatElement(critical[i]))
		}
	}

	if len(high) > 0 {
		remaining := maxElements - len(parts)
		for i := 0; i < len(high) && i < remaining; i++ {
			parts = append(parts, formatElement(high[i]))
		}
	}

	if len(medium) > 0 {
		remaining := maxElements - len(parts)
		for i := 0; i < len(medium) && i < remaining; i++ {
			parts = append(parts, formatElement(medium[i]))
		}
	}

	return strings.Join(parts, "\n")
}

func formatElement(el PageElement) string {
	if len(el.Text) > 100 {
		el.Text = el.Text[:100] + "..."
	}

	// Формируем полный селектор правильно
	var fullSelector string

	if el.Selector == "" {
		// Пустой селектор - используем только тег
		fullSelector = el.Tag
	} else if el.Selector == el.Tag {
		// Селектор совпадает с тегом - используем как есть
		fullSelector = el.Tag
	} else if len(el.Selector) > 0 && (el.Selector[0] == '[' || el.Selector[0] == '#' || el.Selector[0] == '.') {
		// Селектор начинается с спецсимвола: [attr], #id, .class
		// Объединяем с тегом напрямую
		fullSelector = el.Tag + el.Selector
	} else if strings.Contains(el.Selector, el.Tag) {
		// Селектор уже содержит тег (например: "a.class" или "button#id")
		// Используем селектор как есть, без добавления тега
		fullSelector = el.Selector
	} else if strings.HasPrefix(el.Selector, ":") {
		// Псевдо-селектор (:hover, :nth-child)
		fullSelector = el.Tag + el.Selector
	} else {
		// Редкий случай: селектор это что-то другое (может быть имя атрибута)
		// Оборачиваем в квадратные скобки как атрибутный селектор
		fullSelector = el.Tag + "[" + el.Selector + "]"
	}

	return fullSelector + ": " + el.Text
}

func (a *Agent) chunkContext(context string) string {
	words := strings.Fields(context)
	maxWords := a.maxTokens / 2

	if len(words) <= maxWords {
		return context
	}

	return strings.Join(words[:maxWords], " ") + "..."
}
