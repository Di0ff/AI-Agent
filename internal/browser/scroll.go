package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

func (b *PlaywrightBrowser) ScrollToElement(ctx context.Context, selector string) error {
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

	element, err := b.page.QuerySelector(selector)
	if err != nil {
		return fmt.Errorf("элемент не найден: %w", err)
	}

	if element == nil {
		return fmt.Errorf("элемент с селектором %s не найден", selector)
	}

	// Проверяем, виден ли элемент (IsVisible проверяет и видимость, и наличие в DOM)
	isVisible, err := element.IsVisible()
	if err != nil {
		// Если не можем проверить видимость, все равно пытаемся прокрутить
		isVisible = false
	}

	// Если элемент уже виден, не нужно прокручивать
	if isVisible {
		return nil
	}

	// Используем Playwright's встроенный метод прокрутки - более надежный
	// Это предотвращает бесконечную прокрутку и работает синхронно
	err = element.ScrollIntoViewIfNeeded(playwright.ElementHandleScrollIntoViewIfNeededOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		// Если ScrollIntoViewIfNeeded не работает, используем простой scrollIntoView с auto
		_, err = element.Evaluate(`el => {
			el.scrollIntoView({
				behavior: 'auto',
				block: 'center',
				inline: 'center'
			});
		}`)
		if err != nil {
			return fmt.Errorf("ошибка прокрутки к элементу: %w", err)
		}
		// Даем время на завершение прокрутки
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

func (b *PlaywrightBrowser) isElementInViewport(element playwright.ElementHandle) (bool, error) {
	if element == nil {
		return false, fmt.Errorf("element is nil")
	}

	// Используем Evaluate на самом элементе напрямую
	result, err := element.Evaluate(`el => {
		if (!el) return false;

		const rect = el.getBoundingClientRect();
		const windowHeight = window.innerHeight || document.documentElement.clientHeight;
		const windowWidth = window.innerWidth || document.documentElement.clientWidth;

		const vertInView = (rect.top <= windowHeight) && ((rect.top + rect.height) >= 0);
		const horInView = (rect.left <= windowWidth) && ((rect.left + rect.width) >= 0);

		return vertInView && horInView;
	}`)

	if err != nil {
		return false, err
	}

	if inView, ok := result.(bool); ok {
		return inView, nil
	}

	return false, nil
}

func (b *PlaywrightBrowser) ScrollToTop(ctx context.Context) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	_, err := b.page.Evaluate(`() => {
		window.scrollTo({
			top: 0,
			behavior: 'smooth'
		});
	}`)

	if err != nil {
		return fmt.Errorf("ошибка прокрутки наверх: %w", err)
	}

	time.Sleep(300 * time.Millisecond)
	return nil
}

func (b *PlaywrightBrowser) ScrollToBottom(ctx context.Context) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	_, err := b.page.Evaluate(`() => {
		window.scrollTo({
			top: document.body.scrollHeight,
			behavior: 'smooth'
		});
	}`)

	if err != nil {
		return fmt.Errorf("ошибка прокрутки вниз: %w", err)
	}

	time.Sleep(300 * time.Millisecond)
	return nil
}

func (b *PlaywrightBrowser) ScrollByAmount(ctx context.Context, x, y int) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	_, err := b.page.Evaluate(`(coords) => {
		window.scrollBy({
			top: coords.y,
			left: coords.x,
			behavior: 'smooth'
		});
	}`, map[string]int{"x": x, "y": y})

	if err != nil {
		return fmt.Errorf("ошибка прокрутки: %w", err)
	}

	time.Sleep(300 * time.Millisecond)
	return nil
}

