package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

func New(cfg Config) *PlaywrightBrowser {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &PlaywrightBrowser{
		cfg: cfg,
	}
}

func (b *PlaywrightBrowser) Launch(ctx context.Context) error {
	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	b.pw = pw

	opts := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(b.cfg.Headless),
	}

	if b.cfg.UserDataDir != "" {
		userDataDir := "--user-data-dir=" + b.cfg.UserDataDir
		opts.Args = []string{userDataDir}
	}

	if b.cfg.Display != "" {
		opts.Env = map[string]string{
			"DISPLAY": b.cfg.Display,
		}
	}

	browser, err := pw.Chromium.Launch(opts)
	if err != nil {
		return err
	}
	b.browser = browser

	page, err := browser.NewPage()
	if err != nil {
		return err
	}
	b.page = page

	page.SetDefaultTimeout(float64(b.cfg.Timeout.Milliseconds()))

	return nil
}

func (b *PlaywrightBrowser) Navigate(ctx context.Context, url string) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	_, err := b.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return err
	}

	if err := b.ClosePopups(ctx); err != nil {
		return fmt.Errorf("ошибка закрытия попапов после навигации: %w", err)
	}

	return nil
}

func (b *PlaywrightBrowser) Click(ctx context.Context, selector string) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	if err := b.WaitForSelector(ctx, selector); err != nil {
		return fmt.Errorf("элемент не найден: %w", err)
	}

	if err := b.ClosePopups(ctx); err != nil {
		return fmt.Errorf("ошибка закрытия попапов перед кликом: %w", err)
	}

	return b.page.Click(selector)
}

func (b *PlaywrightBrowser) Type(ctx context.Context, selector, text string) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	if err := b.WaitForSelector(ctx, selector); err != nil {
		return fmt.Errorf("элемент не найден: %w", err)
	}

	if err := b.ClosePopups(ctx); err != nil {
		return fmt.Errorf("ошибка закрытия попапов перед вводом: %w", err)
	}

	return b.page.Fill(selector, text)
}

func (b *PlaywrightBrowser) GetPageContext(ctx context.Context) (string, error) {
	if b.page == nil {
		return "", fmt.Errorf("браузер не запущен")
	}

	if err := b.WaitForLoadState(ctx, "networkidle"); err != nil {
		return "", fmt.Errorf("ошибка ожидания загрузки страницы: %w", err)
	}

	content, err := b.page.Content()
	if err != nil {
		return "", err
	}
	return content, nil
}

func (b *PlaywrightBrowser) Close() error {
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
