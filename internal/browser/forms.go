package browser

import (
	"context"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type FormField struct {
	Selector string
	Type     string
	Name     string
	Label    string
	Required bool
	Value    string
}

func (b *PlaywrightBrowser) FindFormFields(ctx context.Context, formSelector string) ([]FormField, error) {
	if b.page == nil {
		return nil, fmt.Errorf("браузер не запущен")
	}

	var form playwright.ElementHandle
	var err error

	if formSelector != "" {
		form, err = b.page.QuerySelector(formSelector)
		if err != nil {
			return nil, fmt.Errorf("форма не найдена по селектору %s: %w", formSelector, err)
		}
	} else {
		form, err = b.page.QuerySelector("form")
		if err != nil {
			return nil, fmt.Errorf("форма не найдена на странице: %w", err)
		}
	}

	if form == nil {
		return []FormField{}, nil
	}

	inputs, err := form.QuerySelectorAll("input, textarea, select")
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска полей формы: %w", err)
	}

	fields := make([]FormField, 0)
	for _, input := range inputs {
		field, err := b.extractFieldInfo(ctx, input)
		if err != nil {
			continue
		}

		if field != nil && field.Selector != "" {
			fields = append(fields, *field)
		}
	}

	return fields, nil
}

func (b *PlaywrightBrowser) extractFieldInfo(ctx context.Context, element playwright.ElementHandle) (*FormField, error) {
	tagName, err := element.Evaluate("el => el.tagName.toLowerCase()")
	if err != nil {
		return nil, err
	}

	tagNameStr := fmt.Sprintf("%v", tagName)

	inputType, _ := element.GetAttribute("type")
	typeStr := inputType
	if typeStr == "" {
		if tagNameStr == "textarea" {
			typeStr = "textarea"
		} else if tagNameStr == "select" {
			typeStr = "select"
		}
	}

	if typeStr == "hidden" || typeStr == "submit" || typeStr == "button" {
		return nil, nil
	}

	id, _ := element.GetAttribute("id")
	name, _ := element.GetAttribute("name")
	placeholder, _ := element.GetAttribute("placeholder")
	ariaLabel, _ := element.GetAttribute("aria-label")
	requiredAttr, _ := element.GetAttribute("required")

	required := requiredAttr != ""
	if !required {
		evalRequired, err := element.Evaluate("el => el.required")
		if err == nil {
			if requiredBool, ok := evalRequired.(bool); ok {
				required = requiredBool
			}
		}
	}

	value, _ := element.GetAttribute("value")
	if value == "" {
		evalValue, err := element.Evaluate("el => el.value || ''")
		if err == nil {
			if valueStr, ok := evalValue.(string); ok {
				value = valueStr
			}
		}
	}

	label := b.findLabelForField(ctx, element, id, placeholder, name, ariaLabel)

	selector := b.buildSelector(id, name, typeStr, tagNameStr)

	field := &FormField{
		Selector: selector,
		Type:     typeStr,
		Name:     name,
		Label:    label,
		Required: required,
		Value:    value,
	}

	return field, nil
}

func (b *PlaywrightBrowser) findLabelForField(ctx context.Context, element playwright.ElementHandle, id, placeholder, name, ariaLabel string) string {
	if ariaLabel != "" {
		return ariaLabel
	}

	if id != "" {
		label, err := b.page.QuerySelector(fmt.Sprintf("label[for='%s']", id))
		if err == nil && label != nil {
			text, err := label.TextContent()
			if err == nil && text != "" {
				return text
			}
		}
	}

	closestLabel, err := element.Evaluate("el => { const label = el.closest('label'); return label ? label.textContent.trim() : null; }")
	if err == nil {
		if labelText, ok := closestLabel.(string); ok && labelText != "" {
			return labelText
		}
	}

	if placeholder != "" {
		return placeholder
	}

	if name != "" {
		return name
	}

	return ""
}

func (b *PlaywrightBrowser) buildSelector(id, name, inputType, tagName string) string {
	if id != "" {
		return "#" + id
	}

	if name != "" {
		return fmt.Sprintf("[name='%s']", name)
	}

	if inputType != "" && tagName == "input" {
		return fmt.Sprintf("input[type='%s']", inputType)
	}

	return tagName
}

func (b *PlaywrightBrowser) FillFormField(ctx context.Context, selector, value string) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	if err := b.WaitForSelector(ctx, selector); err != nil {
		return fmt.Errorf("поле формы не найдено: %w", err)
	}

	element, err := b.page.QuerySelector(selector)
	if err != nil {
		return fmt.Errorf("ошибка получения элемента: %w", err)
	}

	tagName, err := element.Evaluate("el => el.tagName.toLowerCase()")
	if err != nil {
		return fmt.Errorf("ошибка определения типа элемента: %w", err)
	}

	tagNameStr := fmt.Sprintf("%v", tagName)
	if tagNameStr == "select" {
		_, err := b.page.SelectOption(selector, playwright.SelectOptionValues{Values: &[]string{value}})
		return err
	}

	return b.page.Fill(selector, value)
}

func (b *PlaywrightBrowser) SubmitForm(ctx context.Context, formSelector string) error {
	if b.page == nil {
		return fmt.Errorf("браузер не запущен")
	}

	selector := formSelector
	if selector == "" {
		selector = "form"
	}

	if err := b.WaitForSelector(ctx, selector); err != nil {
		return fmt.Errorf("форма не найдена: %w", err)
	}

	submitButton, err := b.page.QuerySelector(selector + " button[type='submit'], " + selector + " input[type='submit'], " + selector + " button:has-text('Отправить'), " + selector + " button:has-text('Submit')")
	if err == nil && submitButton != nil {
		isVisible, _ := submitButton.IsVisible()
		if isVisible {
			return submitButton.Click()
		}
	}

	return b.page.Press(selector+" input, "+selector+" textarea", "Enter")
}

func (b *PlaywrightBrowser) ValidateForm(ctx context.Context, formSelector string) (bool, []string, error) {
	if b.page == nil {
		return false, nil, fmt.Errorf("браузер не запущен")
	}

	var form playwright.ElementHandle
	var err error

	if formSelector != "" {
		form, err = b.page.QuerySelector(formSelector)
	} else {
		form, err = b.page.QuerySelector("form")
	}

	if err != nil || form == nil {
		return false, []string{"Форма не найдена"}, nil
	}

	requiredInputs, err := form.QuerySelectorAll("input[required], textarea[required], select[required]")
	if err != nil {
		return false, nil, fmt.Errorf("ошибка поиска обязательных полей: %w", err)
	}

	errors := []string{}
	for _, input := range requiredInputs {
		value, _ := input.GetAttribute("value")
		if value == "" {
			evalValue, err := input.Evaluate("el => el.value || ''")
			if err == nil {
				if valueStr, ok := evalValue.(string); ok && valueStr == "" {
					name, _ := input.GetAttribute("name")
					id, _ := input.GetAttribute("id")
					fieldName := "поле"
					if name != "" {
						fieldName = name
					} else if id != "" {
						fieldName = id
					}
					errors = append(errors, fmt.Sprintf("Поле \"%s\" обязательно для заполнения", fieldName))
				}
			}
		}
	}

	emailInputs, err := form.QuerySelectorAll("input[type='email']")
	if err == nil {
		for _, emailInput := range emailInputs {
			isValid, err := emailInput.Evaluate("el => el.validity.valid")
			if err == nil {
				if valid, ok := isValid.(bool); ok && !valid {
					value, _ := emailInput.GetAttribute("value")
					if value != "" {
						errors = append(errors, "Некорректный email адрес")
					}
				}
			}
		}
	}

	return len(errors) == 0, errors, nil
}
