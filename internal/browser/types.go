// Package browser предоставляет обертку над Playwright для автоматизации браузера Firefox.
// Пакет включает функции для навигации, работы с элементами, формами, попапами и AJAX запросами.
package browser

import (
	"context"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Browser определяет интерфейс для управления веб-браузером.
// Все операции поддерживают контекст для отмены и таймауты.
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

// PageSnapshot представляет снимок состояния веб-страницы.
// Включает URL, заголовок, список интерактивных элементов и дерево доступности.
type PageSnapshot struct {
	URL               string        // URL страницы
	Title             string        // Заголовок страницы
	Elements          []ElementInfo // Список элементов на странице
	Viewport          ViewportBounds // Размеры viewport
	AccessibilityTree string        // Дерево доступности в JSON формате
}

// ElementInfo содержит информацию об элементе на странице.
type ElementInfo struct {
	Tag         string         // HTML тег элемента
	Text        string         // Текстовое содержимое
	Selector    string         // CSS селектор для доступа
	Visible     bool           // Виден ли элемент
	Interactive bool           // Можно ли взаимодействовать
	InViewport  bool           // Находится ли в видимой области
	Bounds      ViewportBounds // Координаты и размеры
	Role        string         // ARIA роль
	Label       string         // ARIA label
	Priority    int            // Приоритет элемента
}

// ViewportBounds определяет границы области (viewport или элемента).
type ViewportBounds struct {
	X      float64 // X координата
	Y      float64 // Y координата
	Width  float64 // Ширина
	Height float64 // Высота
}

// PlaywrightBrowser реализует интерфейс Browser используя Playwright.
// Поддерживает concurrent доступ через sync.RWMutex.
type PlaywrightBrowser struct {
	pw             *playwright.Playwright
	browser        playwright.Browser
	context        playwright.BrowserContext
	page           playwright.Page
	cfg            Config
	popupDetector  PopupDetector
	mu             sync.RWMutex // Защита от concurrent доступа к page, browser, context
}

// Config содержит конфигурацию для браузера.
type Config struct {
	Headless        bool          // Headless режим (без GUI)
	UserDataDir     string        // Директория для сохранения сессий
	BrowsersPath    string        // Путь к браузерам Playwright
	Display         string        // DISPLAY для Linux (например :0)
	Timeout         time.Duration // Дефолтный timeout для большинства операций
	NavigateTimeout time.Duration // Timeout для navigate операций (обычно больше)
	ActionTimeout   time.Duration // Timeout для click/type операций
}
