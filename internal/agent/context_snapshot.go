package agent

import (
	"fmt"
	"strings"

	"aiAgent/internal/browser"
)

func (a *Agent) limitContextFromSnapshot(snapshot *browser.PageSnapshot) string {
	if snapshot == nil || len(snapshot.Elements) == 0 {
		return ""
	}

	elements := a.convertSnapshotElements(snapshot.Elements)
	elements = a.prioritizeElementsWithViewport(elements, snapshot.Viewport)
	elements = a.filterVisibleAndViewport(elements)
	elements = a.deduplicateElements(elements)

	limited := a.buildLimitedContextFromSnapshot(elements, snapshot)

	if len(limited) <= a.maxTokens {
		return limited
	}

	return a.chunkContext(limited)
}

func (a *Agent) convertSnapshotElements(snapshotElements []browser.ElementInfo) []PageElement {
	elements := make([]PageElement, len(snapshotElements))
	for i, elem := range snapshotElements {
		priority := PriorityMedium
		if elem.Priority >= 5 {
			priority = PriorityCritical
		} else if elem.Priority >= 3 {
			priority = PriorityHigh
		} else if elem.Priority >= 1 {
			priority = PriorityMedium
		} else {
			priority = PriorityLow
		}

		elements[i] = PageElement{
			Tag:         elem.Tag,
			Text:        elem.Text,
			Selector:    elem.Selector,
			Priority:    priority,
			Visible:     elem.Visible,
			Interactive: elem.Interactive,
		}
	}
	return elements
}

func (a *Agent) prioritizeElementsWithViewport(elements []PageElement, viewport browser.ViewportBounds) []PageElement {
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

func (a *Agent) filterVisibleAndViewport(elements []PageElement) []PageElement {
	var filtered []PageElement
	for _, el := range elements {
		if el.Visible {
			filtered = append(filtered, el)
		}
	}
	return filtered
}

func (a *Agent) buildLimitedContextFromSnapshot(elements []PageElement, snapshot *browser.PageSnapshot) string {
	var parts []string

	if snapshot.Title != "" {
		parts = append(parts, fmt.Sprintf("Title: %s", snapshot.Title))
	}

	if snapshot.URL != "" {
		parts = append(parts, fmt.Sprintf("URL: %s", snapshot.URL))
	}

	if snapshot.Viewport.Width > 0 && snapshot.Viewport.Height > 0 {
		parts = append(parts, fmt.Sprintf("Viewport: %.0fx%.0f", snapshot.Viewport.Width, snapshot.Viewport.Height))
	}

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

	if snapshot.AccessibilityTree != "" && len(parts) < maxElements {
		parts = append(parts, "\nAccessibility Tree:\n"+snapshot.AccessibilityTree)
	}

	return strings.Join(parts, "\n")
}
