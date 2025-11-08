package agent

import (
	"regexp"
	"strings"
)

func (a *Agent) extractElements(html string) []PageElement {
	var elements []PageElement

	interactiveTags := map[string]bool{
		"button": true, "a": true, "input": true, "select": true,
		"textarea": true, "form": true, "[role=button]": true,
		"[role=link]": true, "[role=tab]": true,
	}

	// Go regexp не поддерживает обратные ссылки (\1), используем упрощенный подход
	// Находим все открывающие теги (включая самозакрывающиеся)
	tagRegex := regexp.MustCompile(`<(\w+)([^>]*?)(?:>|/>)`)
	matches := tagRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		tag := strings.ToLower(match[1])
		attrs := match[2]

		// Извлекаем текст из атрибутов (aria-label, title, alt) или ищем текст после тега
		text := a.extractTextFromAttributes(attrs)
		if text == "" {
			// Пытаемся найти текст после этого тега до следующего тега
			text = a.extractTextAfterTag(html, match[0])
		}

		if len(text) == 0 && !interactiveTags[tag] {
			continue
		}

		selector := a.buildSelector(tag, attrs)
		interactive := interactiveTags[tag] || strings.Contains(attrs, "onclick") || strings.Contains(attrs, "role=")

		elements = append(elements, PageElement{
			Tag:         tag,
			Text:        text,
			Selector:    selector,
			Priority:    PriorityMedium,
			Visible:     true,
			Interactive: interactive,
		})
	}

	return elements
}

func (a *Agent) extractText(html string) string {
	textRegex := regexp.MustCompile(`>([^<]+)<`)
	matches := textRegex.FindAllStringSubmatch(html, -1)

	var texts []string
	for _, match := range matches {
		if len(match) > 1 {
			text := strings.TrimSpace(match[1])
			if len(text) > 0 {
				texts = append(texts, text)
			}
		}
	}

	return strings.Join(texts, " ")
}

func (a *Agent) extractTextFromAttributes(attrs string) string {
	// Извлекаем текст из aria-label, title, alt атрибутов
	ariaLabelRegex := regexp.MustCompile(`aria-label=["']([^"']+)["']`)
	titleRegex := regexp.MustCompile(`title=["']([^"']+)["']`)
	altRegex := regexp.MustCompile(`alt=["']([^"']+)["']`)

	if match := ariaLabelRegex.FindStringSubmatch(attrs); len(match) > 1 {
		return match[1]
	}
	if match := titleRegex.FindStringSubmatch(attrs); len(match) > 1 {
		return match[1]
	}
	if match := altRegex.FindStringSubmatch(attrs); len(match) > 1 {
		return match[1]
	}
	return ""
}

func (a *Agent) extractTextAfterTag(html, tagMatch string) string {
	// Находим позицию тега в HTML
	idx := strings.Index(html, tagMatch)
	if idx == -1 {
		return ""
	}

	// Ищем текст после закрывающей скобки тега до следующего тега
	start := idx + len(tagMatch)
	if start >= len(html) {
		return ""
	}

	// Ищем текст до следующего < или до конца строки
	end := strings.Index(html[start:], "<")
	if end == -1 {
		end = len(html) - start
	}

	text := strings.TrimSpace(html[start : start+end])
	return text
}

func (a *Agent) buildSelector(tag, attrs string) string {
	idRegex := regexp.MustCompile(`id=["']([^"']+)["']`)
	classRegex := regexp.MustCompile(`class=["']([^"']+)["']`)

	if idMatch := idRegex.FindStringSubmatch(attrs); len(idMatch) > 1 {
		return "#" + idMatch[1]
	}

	if classMatch := classRegex.FindStringSubmatch(attrs); len(classMatch) > 1 {
		classes := strings.Fields(classMatch[1])
		if len(classes) > 0 {
			return tag + "." + classes[0]
		}
	}

	return tag
}
