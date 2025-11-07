package browser

import (
	"context"
	"fmt"

	"aiAgent/internal/extractor"
)

func (b *PlaywrightBrowser) GetPageSnapshot(ctx context.Context) (*PageSnapshot, error) {
	if b.page == nil {
		return nil, fmt.Errorf("браузер не запущен")
	}

	if err := b.WaitForLoadState(ctx, "networkidle"); err != nil {
		return nil, fmt.Errorf("ошибка ожидания загрузки страницы: %w", err)
	}

	snapshot, err := extractor.ExtractPageSnapshot(ctx, b.page)
	if err != nil {
		return nil, fmt.Errorf("ошибка извлечения snapshot: %w", err)
	}

	elements := make([]ElementInfo, len(snapshot.Elements))
	for i, elem := range snapshot.Elements {
		elements[i] = ElementInfo{
			Tag:         elem.Tag,
			Text:        elem.Text,
			Selector:    elem.Selector,
			Visible:     elem.Visible,
			Interactive: elem.Interactive,
			InViewport:  elem.InViewport,
			Bounds: ViewportBounds{
				X:      elem.Bounds.X,
				Y:      elem.Bounds.Y,
				Width:  elem.Bounds.Width,
				Height: elem.Bounds.Height,
			},
			Role:     elem.Role,
			Label:    elem.Label,
			Priority: elem.Priority,
		}
	}

	return &PageSnapshot{
		URL:               snapshot.URL,
		Title:             snapshot.Title,
		Elements:          elements,
		Viewport:          ViewportBounds(snapshot.Viewport),
		AccessibilityTree: snapshot.AccessibilityTree,
	}, nil
}
