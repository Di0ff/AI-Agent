package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func (b *PlaywrightBrowser) WaitForNavigation(ctx context.Context, options ...WaitNavigationOption) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	opts := WaitNavigationOptions{
		Timeout:   b.cfg.Timeout,
		WaitUntil: "networkidle",
	}

	for _, opt := range options {
		opt(&opts)
	}

	return b.WaitForLoadState(ctx, opts.WaitUntil)
}

func (b *PlaywrightBrowser) WaitForRequest(ctx context.Context, urlPattern string, timeout time.Duration) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	if timeout == 0 {
		timeout = b.cfg.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ch := make(chan error, 1)
	done := make(chan struct{})

	b.page.OnRequest(func(request playwright.Request) {
		select {
		case <-done:
			return
		default:
		}
		url := request.URL()
		if strings.Contains(url, urlPattern) {
			close(done)
			ch <- nil
		}
	})

	select {
	case <-ctx.Done():
		return fmt.Errorf("таймаут ожидания запроса с паттерном %s", urlPattern)
	case err := <-ch:
		return err
	}
}

func (b *PlaywrightBrowser) WaitForResponse(ctx context.Context, urlPattern string, timeout time.Duration) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	if timeout == 0 {
		timeout = b.cfg.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ch := make(chan error, 1)
	done := make(chan struct{})

	b.page.OnResponse(func(response playwright.Response) {
		select {
		case <-done:
			return
		default:
		}
		url := response.URL()
		if strings.Contains(url, urlPattern) {
			close(done)
			if response.Status() >= 200 && response.Status() < 400 {
				ch <- nil
			} else {
				ch <- fmt.Errorf("ответ с ошибкой: статус %d", response.Status())
			}
		}
	})

	select {
	case <-ctx.Done():
		return fmt.Errorf("таймаут ожидания ответа с паттерном %s", urlPattern)
	case err := <-ch:
		return err
	}
}

func (b *PlaywrightBrowser) WaitForNetworkIdle(ctx context.Context, timeout time.Duration) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	if timeout == 0 {
		timeout = b.cfg.Timeout
	}

	return b.WaitForLoadState(ctx, "networkidle")
}

type WaitNavigationOptions struct {
	Timeout   time.Duration
	WaitUntil string
}

type WaitNavigationOption func(*WaitNavigationOptions)

func WithNavigationTimeout(timeout time.Duration) WaitNavigationOption {
	return func(opts *WaitNavigationOptions) {
		opts.Timeout = timeout
	}
}

func WithNavigationWaitUntil(waitUntil string) WaitNavigationOption {
	return func(opts *WaitNavigationOptions) {
		opts.WaitUntil = waitUntil
	}
}
