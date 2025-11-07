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
	GetPageSnapshot(ctx context.Context) (*PageSnapshot, error)
	WaitForSelector(ctx context.Context, selector string) error
	WaitForLoadState(ctx context.Context, state string) error
	ClosePopups(ctx context.Context) error
	FindFormFields(ctx context.Context, formSelector string) ([]FormField, error)
	FillFormField(ctx context.Context, selector, value string) error
	SubmitForm(ctx context.Context, formSelector string) error
	ValidateForm(ctx context.Context, formSelector string) (bool, []string, error)
	WaitForNavigation(ctx context.Context, options ...WaitNavigationOption) error
	WaitForRequest(ctx context.Context, urlPattern string, timeout time.Duration) error
	WaitForResponse(ctx context.Context, urlPattern string, timeout time.Duration) error
	WaitForNetworkIdle(ctx context.Context, timeout time.Duration) error
	Close() error
}

type PageSnapshot struct {
	URL               string
	Title             string
	Elements          []ElementInfo
	Viewport          ViewportBounds
	AccessibilityTree string
}

type ElementInfo struct {
	Tag         string
	Text        string
	Selector    string
	Visible     bool
	Interactive bool
	InViewport  bool
	Bounds      ViewportBounds
	Role        string
	Label       string
	Priority    int
}

type ViewportBounds struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type PlaywrightBrowser struct {
	pw             *playwright.Playwright
	browser        playwright.Browser
	page           playwright.Page
	cfg            Config
	popupDetector  PopupDetector
}

type Config struct {
	Headless     bool
	UserDataDir  string
	BrowsersPath string
	Display      string
	Timeout      time.Duration
}
