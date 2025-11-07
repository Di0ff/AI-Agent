package browser

import (
	"context"
	"fmt"
	"time"
)

func (b *PlaywrightBrowser) ScrollToElement(ctx context.Context, selector string) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	element, err := b.page.QuerySelector(selector)
	if err != nil {
		return fmt.Errorf("элемент не найден: %w", err)
	}

	if element == nil {
		return fmt.Errorf("элемент с селектором %s не найден", selector)
	}

	isVisible, err := element.IsVisible()
	if err != nil {
		return fmt.Errorf("ошибка проверки видимости: %w", err)
	}

	if !isVisible {
		_, err = element.Evaluate(`el => {
			el.scrollIntoView({
				behavior: 'smooth',
				block: 'center',
				inline: 'center'
			});
		}`)
		if err != nil {
			return fmt.Errorf("ошибка прокрутки к элементу: %w", err)
		}

		time.Sleep(300 * time.Millisecond)
	}

	inViewport, err := b.isElementInViewport(element)
	if err != nil {
		return fmt.Errorf("ошибка проверки viewport: %w", err)
	}

	if !inViewport {
		_, err = element.Evaluate(`el => {
			el.scrollIntoView({
				behavior: 'smooth',
				block: 'center',
				inline: 'center'
			});
		}`)
		if err != nil {
			return fmt.Errorf("ошибка прокрутки к элементу: %w", err)
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (b *PlaywrightBrowser) isElementInViewport(element interface{}) (bool, error) {
	result, err := b.page.Evaluate(`(selector) => {
		const el = document.querySelector(selector);
		if (!el) return false;

		const rect = el.getBoundingClientRect();
		const windowHeight = window.innerHeight || document.documentElement.clientHeight;
		const windowWidth = window.innerWidth || document.documentElement.clientWidth;

		const vertInView = (rect.top <= windowHeight) && ((rect.top + rect.height) >= 0);
		const horInView = (rect.left <= windowWidth) && ((rect.left + rect.width) >= 0);

		return vertInView && horInView;
	}`, element)

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
