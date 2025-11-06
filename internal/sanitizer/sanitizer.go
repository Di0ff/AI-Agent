package sanitizer

import (
	"regexp"
	"strings"
)

type DataSanitizer struct {
	rules []SanitizerRule
}

type SanitizerRule interface {
	Sanitize(text string) string
}

func New() *DataSanitizer {
	return &DataSanitizer{
		rules: []SanitizerRule{
			&PasswordSanitizer{},
			&TokenSanitizer{},
			&CookieSanitizer{},
			&CardSanitizer{},
			&APIKeySanitizer{},
			&EmailSanitizer{},
			&PhoneSanitizer{},
			&AddressSanitizer{},
		},
	}
}

func NewWithAI(llmClient LLMClient) *DataSanitizer {
	rules := []SanitizerRule{
		&PasswordSanitizer{},
		&TokenSanitizer{},
		&CookieSanitizer{},
		&CardSanitizer{},
		&APIKeySanitizer{},
		&EmailSanitizer{},
		&PhoneSanitizer{},
		&AddressSanitizer{},
	}

	if llmClient != nil {
		rules = append(rules, NewAISanitizerRule(llmClient))
	}

	return &DataSanitizer{
		rules: rules,
	}
}

func (s *DataSanitizer) Sanitize(text string) string {
	if text == "" {
		return text
	}

	result := text
	for _, rule := range s.rules {
		result = rule.Sanitize(result)
	}

	return result
}

func (s *DataSanitizer) SanitizeSelector(selector string) string {
	if selector == "" {
		return selector
	}

	lower := strings.ToLower(selector)

	sensitiveKeywords := []string{
		"password", "пароль", "token", "api-key", "api_key",
		"email", "phone", "телефон", "address", "адрес",
	}

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lower, keyword) {
			return "[FILTERED_SELECTOR]"
		}
	}

	return selector
}

func (s *DataSanitizer) SanitizeValue(value string) string {
	if value == "" {
		return value
	}

	if len(value) > 50 {
		return s.Sanitize(value)
	}

	if s.looksLikeSensitiveData(value) {
		return "[FILTERED]"
	}

	return s.Sanitize(value)
}

func (s *DataSanitizer) looksLikeSensitiveData(value string) bool {
	lower := strings.ToLower(value)

	sensitivePatterns := []string{
		"password", "пароль", "token", "api", "secret",
		"card", "cvv", "cvc", "expir", "session",
		"email", "phone", "телефон", "address", "адрес",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	if len(value) > 20 && regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(value) {
		return true
	}

	return false
}
