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
	return el.Tag + "[" + el.Selector + "]: " + el.Text
}

func (a *Agent) chunkContext(context string) string {
	words := strings.Fields(context)
	maxWords := a.maxTokens / 2

	if len(words) <= maxWords {
		return context
	}

	return strings.Join(words[:maxWords], " ") + "..."
}
