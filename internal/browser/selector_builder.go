package browser

import (
	"fmt"
	"regexp"
	"strings"
)

type SelectorStrategy int

const (
	StrategyID SelectorStrategy = iota
	StrategyDataTestID
	StrategyName
	StrategyAriaLabel
	StrategyClass
	StrategyXPath
	StrategyNthChild
	StrategyTag
)

type SmartSelector struct {
	Selector string
	Strategy SelectorStrategy
	Score    int
}

func BuildSmartSelector(element map[string]interface{}) string {
	candidates := []SmartSelector{}

	id := getStringAttr(element, "id")
	if id != "" && isUniqueID(id) {
		candidates = append(candidates, SmartSelector{
			Selector: "#" + id,
			Strategy: StrategyID,
			Score:    100,
		})
	}

	dataTestId := getStringAttr(element, "data-testid")
	if dataTestId != "" {
		candidates = append(candidates, SmartSelector{
			Selector: "[data-testid='" + dataTestId + "']",
			Strategy: StrategyDataTestID,
			Score:    95,
		})
	}

	name := getStringAttr(element, "name")
	if name != "" {
		candidates = append(candidates, SmartSelector{
			Selector: "[name='" + name + "']",
			Strategy: StrategyName,
			Score:    90,
		})
	}

	ariaLabel := getStringAttr(element, "aria-label")
	role := getStringAttr(element, "role")
	if ariaLabel != "" && role != "" {
		candidates = append(candidates, SmartSelector{
			Selector: "[role='" + role + "'][aria-label='" + ariaLabel + "']",
			Strategy: StrategyAriaLabel,
			Score:    85,
		})
	}

	className := getStringAttr(element, "class")
	tag := getStringAttr(element, "tag")
	if className != "" && tag != "" {
		classes := strings.Fields(className)
		if len(classes) > 0 {
			firstClass := classes[0]
			if !isCommonClass(firstClass) {
				candidates = append(candidates, SmartSelector{
					Selector: tag + "." + firstClass,
					Strategy: StrategyClass,
					Score:    70,
				})
			}
		}
	}

	xpath := buildXPathSelector(element)
	if xpath != "" {
		candidates = append(candidates, SmartSelector{
			Selector: xpath,
			Strategy: StrategyXPath,
			Score:    60,
		})
	}

	if tag != "" {
		nthChild := getIntAttr(element, "nth-child")
		if nthChild > 0 {
			parentSelector := getStringAttr(element, "parent-selector")
			if parentSelector != "" {
				candidates = append(candidates, SmartSelector{
					Selector: parentSelector + " > " + tag + ":nth-child(" + fmt.Sprintf("%d", nthChild) + ")",
					Strategy: StrategyNthChild,
					Score:    50,
				})
			} else {
				candidates = append(candidates, SmartSelector{
					Selector: tag + ":nth-child(" + fmt.Sprintf("%d", nthChild) + ")",
					Strategy: StrategyNthChild,
					Score:    40,
				})
			}
		} else {
			candidates = append(candidates, SmartSelector{
				Selector: tag,
				Strategy: StrategyTag,
				Score:    30,
			})
		}
	}

	if len(candidates) == 0 {
		return "body"
	}

	bestSelector := candidates[0]
	for _, candidate := range candidates {
		if candidate.Score > bestSelector.Score {
			bestSelector = candidate
		}
	}

	return bestSelector.Selector
}

func buildXPathSelector(element map[string]interface{}) string {
	tag := getStringAttr(element, "tag")
	if tag == "" {
		return ""
	}

	id := getStringAttr(element, "id")
	if id != "" {
		return "//" + tag + "[@id='" + id + "']"
	}

	text := getStringAttr(element, "text")
	if text != "" && len(text) < 50 {
		return "//" + tag + "[contains(text(), '" + escapeXPath(text) + "')]"
	}

	ariaLabel := getStringAttr(element, "aria-label")
	if ariaLabel != "" {
		return "//" + tag + "[@aria-label='" + escapeXPath(ariaLabel) + "']"
	}

	return ""
}

func getStringAttr(element map[string]interface{}, key string) string {
	if val, ok := element[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntAttr(element map[string]interface{}, key string) int {
	if val, ok := element[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
		if num, ok := val.(int); ok {
			return num
		}
	}
	return 0
}

func isUniqueID(id string) bool {
	if id == "" {
		return false
	}

	commonIDs := []string{"content", "main", "header", "footer", "nav", "menu"}
	idLower := strings.ToLower(id)
	for _, common := range commonIDs {
		if idLower == common {
			return false
		}
	}

	return true
}

func isCommonClass(class string) bool {
	commonClasses := []string{
		"container", "wrapper", "row", "col", "btn", "button",
		"active", "disabled", "hidden", "visible", "flex", "grid",
	}
	classLower := strings.ToLower(class)
	for _, common := range commonClasses {
		if strings.Contains(classLower, common) {
			return true
		}
	}
	return false
}

func escapeXPath(text string) string {
	text = strings.ReplaceAll(text, "'", "\\'")
	text = strings.ReplaceAll(text, "\"", "\\\"")
	return text
}

// NormalizeSelector нормализует селектор, преобразуя невалидные синтаксисы в валидные для Playwright.
// Преобразует jQuery :contains() в Playwright :has-text().
// Также исправляет селекторы вида "button: Текст" в "button:has-text('Текст')".
// Возвращает нормализованный селектор и флаг, указывающий, был ли селектор изменен.
func NormalizeSelector(selector string) (string, bool) {
	if selector == "" {
		return selector, false
	}

	normalized := selector
	changed := false

	// Исправляем селекторы вида "tag: Текст" или "tag.class: Текст" в "tag:has-text('Текст')"
	// Это распространенная ошибка LLM - использование двоеточия с пробелом вместо :has-text()
	// Проверяем паттерн: что-то до двоеточия, затем пробел, затем текст
	colonSpacePattern := regexp.MustCompile(`^([^:]+):\s+(.+)$`)
	if colonSpacePattern.MatchString(normalized) {
		submatch := colonSpacePattern.FindStringSubmatch(normalized)
		if len(submatch) >= 3 {
			tagPart := strings.TrimSpace(submatch[1])
			textPart := strings.TrimSpace(submatch[2])
			
			// Проверяем, что это не валидный псевдокласс CSS (например, :hover, :focus, :has-text)
			// Валидные псевдоклассы не имеют пробела после двоеточия
			validPseudoClasses := []string{":hover", ":focus", ":active", ":visited", ":link", ":checked", 
				":disabled", ":enabled", ":first-child", ":last-child", ":nth-child", ":nth-of-type", 
				":has-text", ":has", ":not", ":contains"}
			isValidPseudo := false
			for _, pseudo := range validPseudoClasses {
				if strings.HasSuffix(tagPart, pseudo) || strings.Contains(normalized, pseudo+"(") {
					isValidPseudo = true
					break
				}
			}
			
			// Если это не валидный псевдокласс, преобразуем в :has-text()
			if !isValidPseudo && tagPart != "" && textPart != "" {
				changed = true
				// Экранируем кавычки в тексте
				textPart = strings.ReplaceAll(textPart, `"`, `\"`)
				textPart = strings.ReplaceAll(textPart, `'`, `\'`)
				normalized = tagPart + `:has-text("` + textPart + `")`
			}
		}
	}

	// Регулярное выражение для поиска :contains("text") с двойными кавычками
	containsPatternDouble := regexp.MustCompile(`:contains\("([^"]*)"\)`)
	normalized = containsPatternDouble.ReplaceAllStringFunc(normalized, func(match string) string {
		changed = true
		submatch := containsPatternDouble.FindStringSubmatch(match)
		if len(submatch) >= 2 {
			text := submatch[1]
			// Экранируем обратные слеши и кавычки в тексте для Playwright
			text = strings.ReplaceAll(text, "\\", "\\\\")
			text = strings.ReplaceAll(text, `"`, `\"`)
			return `:has-text("` + text + `")`
		}
		return match
	})

	// Регулярное выражение для поиска :contains('text') с одинарными кавычками
	containsPatternSingle := regexp.MustCompile(`:contains\('([^']*)'\)`)
	normalized = containsPatternSingle.ReplaceAllStringFunc(normalized, func(match string) string {
		changed = true
		submatch := containsPatternSingle.FindStringSubmatch(match)
		if len(submatch) >= 2 {
			text := submatch[1]
			// Экранируем обратные слеши и кавычки в тексте для Playwright
			text = strings.ReplaceAll(text, "\\", "\\\\")
			text = strings.ReplaceAll(text, `'`, `\'`)
			return `:has-text('` + text + `')`
		}
		return match
	})

	// Также обрабатываем случай без кавычек (редко, но возможно)
	containsPatternNoQuotes := regexp.MustCompile(`:contains\(([^)]+)\)`)
	normalized = containsPatternNoQuotes.ReplaceAllStringFunc(normalized, func(match string) string {
		if !strings.Contains(match, `:has-text(`) {
			changed = true
			submatch := containsPatternNoQuotes.FindStringSubmatch(match)
			if len(submatch) >= 2 {
				text := submatch[1]
				// Если текст не в кавычках, добавляем кавычки
				text = strings.TrimSpace(text)
				return `:has-text("` + text + `")`
			}
		}
		return match
	})

	return normalized, changed
}

// ValidateSelector проверяет, что селектор является валидным CSS/Playwright селектором.
// Возвращает ошибку, если селектор является URL или невалидным.
func ValidateSelector(selector string) error {
	if selector == "" {
		return fmt.Errorf("селектор не может быть пустым")
	}

	// Проверяем, что селектор не является URL
	selectorTrimmed := strings.TrimSpace(selector)
	if strings.HasPrefix(selectorTrimmed, "http://") || strings.HasPrefix(selectorTrimmed, "https://") {
		return fmt.Errorf("селектор не может быть URL. Для перехода по URL используй действие 'navigate', а не 'click'. Получен URL: %s", selector)
	}

	// Проверяем, что селектор не начинается с протокола (ftp://, file:// и т.д.)
	if strings.Contains(selectorTrimmed, "://") {
		return fmt.Errorf("селектор не может содержать протокол (://). Получен: %s", selector)
	}

	// Не проверяем формат "tag: Текст" здесь, так как NormalizeSelector исправит это
	// Валидация должна быть минимальной и проверять только критические ошибки

	return nil
}
