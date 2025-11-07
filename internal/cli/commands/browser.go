package commands

import (
	"context"
	"fmt"
	"strings"

	"aiAgent/internal/browser"
	"aiAgent/internal/cli/ui"
)

// BrowserHandler обрабатывает команды браузера
type BrowserHandler struct {
	browser browser.Browser
	readLine func() (string, error)
}

func NewBrowserHandler(br browser.Browser, readLine func() (string, error)) *BrowserHandler {
	return &BrowserHandler{
		browser: br,
		readLine: readLine,
	}
}

// OpenPersistent открывает браузер в persistent режиме
func (h *BrowserHandler) OpenPersistent(ctx context.Context) {
	if h.browser == nil {
		fmt.Println(ui.ColorRed + ui.IconCross + " Браузер не инициализирован" + ui.ColorReset)
		return
	}

	fmt.Println(ui.ColorCyan + ui.IconGlobe + " Запуск браузера в persistent режиме..." + ui.ColorReset)
	if err := h.browser.Launch(ctx); err != nil {
		fmt.Printf(ui.ColorRed+ui.IconCross+" Ошибка запуска:"+ui.ColorReset+" %v\n", err)
		return
	}

	fmt.Println(ui.ColorGreen + ui.IconCheckmark + " Браузер открыт с сохранением сессии" + ui.ColorReset)
	fmt.Println(ui.ColorGray + "Вы можете вручную залогиниться и настроить браузер" + ui.ColorReset)
	fmt.Println(ui.ColorGray + "Затем используйте команду '" + ui.ColorYellow + "run <id>" + ui.ColorGray + "' для выполнения задач" + ui.ColorReset)
	fmt.Println(ui.ColorYellow + "⏎ Нажмите Enter для закрытия браузера..." + ui.ColorReset)
	h.readLine()
	h.browser.Close()
	fmt.Println(ui.ColorGreen + ui.IconCheckmark + " Сессия сохранена в ./browser-data" + ui.ColorReset)
}

// Open открывает URL в браузере
func (h *BrowserHandler) Open(ctx context.Context, url string) {
	if h.browser == nil {
		fmt.Println(ui.ColorRed + ui.IconCross + " Браузер не инициализирован" + ui.ColorReset)
		return
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	fmt.Println(ui.ColorCyan + ui.IconGlobe + " Запуск браузера..." + ui.ColorReset)
	if err := h.browser.Launch(ctx); err != nil {
		fmt.Printf(ui.ColorRed+ui.IconCross+" Ошибка запуска:"+ui.ColorReset+" %v\n", err)
		return
	}

	fmt.Printf(ui.ColorCyan+ui.IconArrow+" Открытие %s..."+ui.ColorReset+"\n", url)
	if err := h.browser.Navigate(ctx, url); err != nil {
		fmt.Printf(ui.ColorRed+ui.IconCross+" Ошибка навигации:"+ui.ColorReset+" %v\n", err)
		h.browser.Close()
		return
	}

	fmt.Println(ui.ColorGreen + ui.IconCheckmark + " Страница открыта" + ui.ColorReset)
	fmt.Println(ui.ColorYellow + "⏎ Нажмите Enter для закрытия браузера..." + ui.ColorReset)
	h.readLine()
	h.browser.Close()
	fmt.Println(ui.ColorGray + "Браузер закрыт" + ui.ColorReset)
}
