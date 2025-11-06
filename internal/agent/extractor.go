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

	tagRegex := regexp.MustCompile(`<(\w+)([^>]*)>(.*?)</\1>`)
	matches := tagRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		tag := strings.ToLower(match[1])
		attrs := match[2]
		text := a.extractText(match[3])

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
