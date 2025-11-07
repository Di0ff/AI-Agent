package agent

import "strings"

func (a *Agent) prioritizeElements(elements []PageElement) []PageElement {
	for i := range elements {
		if elements[i].Interactive {
			elements[i].Priority = PriorityHigh
		}

		semanticTags := []string{"button", "a", "input", "select", "textarea", "nav", "form"}
		for _, tag := range semanticTags {
			if strings.EqualFold(elements[i].Tag, tag) {
				if elements[i].Priority < PriorityHigh {
					elements[i].Priority = PriorityHigh
				}
				break
			}
		}

		if elements[i].Tag == "a" && len(elements[i].Text) > 0 {
			elements[i].Priority = PriorityHigh
		}
	}

	return elements
}

func (a *Agent) filterVisible(elements []PageElement) []PageElement {
	var visible []PageElement
	for _, el := range elements {
		if el.Visible {
			visible = append(visible, el)
		}
	}
	return visible
}

func (a *Agent) deduplicateElements(elements []PageElement) []PageElement {
	seen := make(map[string]bool)
	var unique []PageElement

	for _, el := range elements {
		key := el.Selector + ":" + el.Text
		if !seen[key] && len(el.Text) > 0 {
			seen[key] = true
			unique = append(unique, el)
		}
	}

	return unique
}
