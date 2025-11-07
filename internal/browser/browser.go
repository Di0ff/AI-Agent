package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

func New(cfg Config) *PlaywrightBrowser {
	// Установка дефолтных таймаутов
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.NavigateTimeout == 0 {
		cfg.NavigateTimeout = 60 * time.Second // Navigate обычно дольше
	}
	if cfg.ActionTimeout == 0 {
		cfg.ActionTimeout = 10 * time.Second // Click/Type обычно быстрые
	}

	return &PlaywrightBrowser{
		cfg: cfg,
	}
}

func (b *PlaywrightBrowser) SetPopupDetector(detector PopupDetector) {
	b.popupDetector = detector
}

// getPage безопасно возвращает текущую страницу с read lock
func (b *PlaywrightBrowser) getPage() playwright.Page {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.page
}

// setPage безопасно устанавливает страницу с write lock
func (b *PlaywrightBrowser) setPage(page playwright.Page) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.page = page
}

func (b *PlaywrightBrowser) getBrowserArgs() []string {
	return []string{
		"--no-sandbox",
	}
}

func (b *PlaywrightBrowser) getEnvMap() map[string]string {
	if b.cfg.Display != "" {
		return map[string]string{
			"DISPLAY": b.cfg.Display,
		}
	}
	return nil
}

func (b *PlaywrightBrowser) launchPersistent(pw *playwright.Playwright) error {
	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: playwright.Bool(b.cfg.Headless),
		Args:     b.getBrowserArgs(),
	}

	if env := b.getEnvMap(); env != nil {
		opts.Env = env
	}

	browserContext, err := pw.Firefox.LaunchPersistentContext(b.cfg.UserDataDir, opts)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.context = browserContext
	b.mu.Unlock()

	pages := browserContext.Pages()
	var page playwright.Page
	if len(pages) == 0 {
		page, err = browserContext.NewPage()
		if err != nil {
			return err
		}
	} else {
		page = pages[0]
	}

	b.setPage(page)
	page.SetDefaultTimeout(float64(b.cfg.Timeout.Milliseconds()))
	return nil
}

func (b *PlaywrightBrowser) launchStandard(pw *playwright.Playwright) error {
	opts := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(b.cfg.Headless),
		Args:     b.getBrowserArgs(),
	}

	if env := b.getEnvMap(); env != nil {
		opts.Env = env
	}

	browser, err := pw.Firefox.Launch(opts)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.browser = browser
	b.mu.Unlock()

	page, err := browser.NewPage()
	if err != nil {
		return err
	}

	b.setPage(page)
	page.SetDefaultTimeout(float64(b.cfg.Timeout.Milliseconds()))
	return nil
}

func (b *PlaywrightBrowser) Launch(ctx context.Context) error {
	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	b.pw = pw

	if b.cfg.UserDataDir != "" {
		return b.launchPersistent(pw)
	}

	return b.launchStandard(pw)
}

func (b *PlaywrightBrowser) Navigate(ctx context.Context, url string) error {
	page := b.getPage()
	if page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	// Создаем context с timeout для navigate операции
	navCtx, cancel := context.WithTimeout(ctx, b.cfg.NavigateTimeout)
	defer cancel()

	// Channel для получения результата
	errChan := make(chan error, 1)
	go func() {
		_, err := page.Goto(url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(float64(b.cfg.NavigateTimeout.Milliseconds())),
		})
		errChan <- err
	}()

	// Ждем результат или timeout
	select {
	case <-navCtx.Done():
		return fmt.Errorf("navigate timeout after %v", b.cfg.NavigateTimeout)
	case err := <-errChan:
		if err != nil {
			return err
		}
	}

	if err := b.ClosePopups(ctx); err != nil {
		return fmt.Errorf("ошибка закрытия попапов после навигации: %w", err)
	}

	return nil
}

func (b *PlaywrightBrowser) Click(ctx context.Context, selector string) error {
	page := b.getPage()
	if page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	if err := b.WaitForSelector(ctx, selector); err != nil {
		return fmt.Errorf("элемент не найден: %w", err)
	}

	if err := b.ClosePopups(ctx); err != nil {
		return fmt.Errorf("ошибка закрытия попапов перед кликом: %w", err)
	}

	if err := b.ScrollToElement(ctx, selector); err != nil {
		return fmt.Errorf("ошибка прокрутки к элементу: %w", err)
	}

	err := page.Click(selector)
	if err != nil {
		return err
	}

	if err := b.WaitForNetworkIdle(ctx, 2*time.Second); err != nil {
		return nil
	}

	return nil
}

func (b *PlaywrightBrowser) Type(ctx context.Context, selector, text string) error {
	page := b.getPage()
	if page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	if err := b.WaitForSelector(ctx, selector); err != nil {
		return fmt.Errorf("элемент не найден: %w", err)
	}

	if err := b.ClosePopups(ctx); err != nil {
		return fmt.Errorf("ошибка закрытия попапов перед вводом: %w", err)
	}

	if err := b.ScrollToElement(ctx, selector); err != nil {
		return fmt.Errorf("ошибка прокрутки к элементу: %w", err)
	}

	return page.Fill(selector, text)
}

func (b *PlaywrightBrowser) GetPageContext(ctx context.Context) (string, error) {
	page := b.getPage()
	if page == nil {
		return "", fmt.Errorf("браузер не запущен")
	}

	if err := b.WaitForLoadState(ctx, "networkidle"); err != nil {
		return "", fmt.Errorf("ошибка ожидания загрузки страницы: %w", err)
	}

	content, err := page.Content()
	if err != nil {
		return "", err
	}
	return content, nil
}

func (b *PlaywrightBrowser) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.context != nil {
		if err := b.context.Close(); err != nil {
			return err
		}
	}
	if b.browser != nil {
		if err := b.browser.Close(); err != nil {
			return err
		}
	}
	if b.pw != nil {
		return b.pw.Stop()
	}
	return nil
}
