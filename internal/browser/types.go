package browser

import (
	"context"
	"time"

	"github.com/playwright-community/playwright-go"
)

type Browser interface {
	Launch(ctx context.Context) error
	Navigate(ctx context.Context, url string) error
	Click(ctx context.Context, selector string) error
	Type(ctx context.Context, selector, text string) error
	GetPageContext(ctx context.Context) (string, error)
	WaitForSelector(ctx context.Context, selector string) error
	WaitForLoadState(ctx context.Context, state string) error
	ClosePopups(ctx context.Context) error
	Close() error
}

type PlaywrightBrowser struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	page    playwright.Page
	cfg     Config
}

type Config struct {
	Headless     bool
	UserDataDir  string
	BrowsersPath string
	Display      string
	Timeout      time.Duration
}
