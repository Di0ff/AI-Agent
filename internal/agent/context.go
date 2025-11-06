package agent

func (a *Agent) limitContext(context string) string {
	if len(context) == 0 {
		return context
	}

	elements := a.extractElements(context)
	elements = a.prioritizeElements(elements)
	elements = a.filterVisible(elements)
	elements = a.deduplicateElements(elements)

	limited := a.buildLimitedContext(elements)

	if len(limited) <= a.maxTokens {
		return limited
	}

	return a.chunkContext(limited)
}
