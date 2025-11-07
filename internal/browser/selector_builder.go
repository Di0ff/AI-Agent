package browser

import (
	"fmt"
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
