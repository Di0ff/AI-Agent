package sanitizer

import "regexp"

type CardSanitizer struct{}

func (s *CardSanitizer) Sanitize(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),
		regexp.MustCompile(`(?i)(card[_-]?number|номер[_-]?карты)\s*[:=]\s*["']?(\d{13,19})["']?`),
		regexp.MustCompile(`(?i)(cvv|cvc)\s*[:=]\s*["']?(\d{3,4})["']?`),
		regexp.MustCompile(`(?i)(cvv2|cvc2)\s*[:=]\s*["']?(\d{3,4})["']?`),
		regexp.MustCompile(`(?i)(expir|срок)\s*[:=]\s*["']?(\d{2}[/-]\d{2,4})["']?`),
	}

	for _, pattern := range patterns {
		text = pattern.ReplaceAllString(text, `[FILTERED]`)
	}

	return text
}
