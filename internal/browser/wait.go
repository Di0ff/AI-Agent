package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func (b *PlaywrightBrowser) WaitForSelector(ctx context.Context, selector string) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	// Валидируем селектор (проверяем, что это не URL)
	if err := ValidateSelector(selector); err != nil {
		return fmt.Errorf("невалидный селектор: %w", err)
	}

	// Нормализуем селектор (преобразуем :contains() в :has-text())
	normalizedSelector, changed := NormalizeSelector(selector)
	if changed {
		// Селектор был изменен, используем нормализованную версию
		selector = normalizedSelector
	}

	opts := playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(b.cfg.Timeout.Seconds() * 1000),
	}

	_, err := b.page.WaitForSelector(selector, opts)
	return err
}

func (b *PlaywrightBrowser) WaitForLoadState(ctx context.Context, state string) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	var loadState *playwright.LoadState
	switch strings.ToLower(state) {
	case "load":
		loadState = playwright.LoadStateLoad
	case "domcontentloaded":
		loadState = playwright.LoadStateDomcontentloaded
	case "networkidle":
		loadState = playwright.LoadStateNetworkidle
	default:
		loadState = playwright.LoadStateLoad
	}

	opts := playwright.PageWaitForLoadStateOptions{
		State:   loadState,
		Timeout: playwright.Float(b.cfg.Timeout.Seconds() * 1000),
	}

	return b.page.WaitForLoadState(opts)
}

func (b *PlaywrightBrowser) ClosePopups(ctx context.Context) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	if b.popupDetector == nil {
		return b.closePopupsLegacy(ctx)
	}

	snapshot, err := b.GetPageSnapshot(ctx)
	if err != nil {
		return b.closePopupsLegacy(ctx)
	}

	popupInfo, err := b.popupDetector.DetectPopup(ctx, snapshot)
	if err != nil {
		return b.closePopupsLegacy(ctx)
	}

	if !popupInfo.HasPopup || popupInfo.CloseSelector == "" {
		return nil
	}

	element, err := b.page.QuerySelector(popupInfo.CloseSelector)
	if err != nil || element == nil {
		return nil
	}

	isVisible, err := element.IsVisible()
	if err != nil || !isVisible {
		return nil
	}

	if err := element.Click(); err == nil {
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (b *PlaywrightBrowser) closePopupsLegacy(ctx context.Context) error {
	popupSelectors := []string{
		"[role='dialog'] button[aria-label*='close' i]",
		"[role='dialog'] button[aria-label*='закрыть' i]",
		".modal button.close",
		".popup button.close",
		"[data-dismiss='modal']",
		".close-button",
		"button:has-text('×')",
		"button:has-text('✕')",
		"[aria-label='Close']",
		"[aria-label='Закрыть']",
	}

	for _, selector := range popupSelectors {
		elements, err := b.page.QuerySelectorAll(selector)
		if err != nil {
			continue
		}

		for _, element := range elements {
			isVisible, err := element.IsVisible()
			if err != nil || !isVisible {
				continue
			}

			if err := element.Click(); err == nil {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	overlaySelectors := []string{
		"[role='dialog']",
		".modal",
		".popup",
		".overlay",
		"[class*='modal']",
		"[class*='popup']",
		"[class*='overlay']",
	}

	for _, selector := range overlaySelectors {
		elements, err := b.page.QuerySelectorAll(selector)
		if err != nil {
			continue
		}

		for _, element := range elements {
			isVisible, err := element.IsVisible()
			if err != nil || !isVisible {
				continue
			}

			closeButton, err := element.QuerySelector("button[aria-label*='close' i], button[aria-label*='закрыть' i], .close, [data-dismiss]")
			if err == nil && closeButton != nil {
				if err := closeButton.Click(); err == nil {
					time.Sleep(500 * time.Millisecond)
				}
			}
		}
	}

	return nil
}
