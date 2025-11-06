package sanitizer

import "regexp"

type PhoneSanitizer struct{}

func (s *PhoneSanitizer) Sanitize(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\+7\s?\(?\d{3}\)?\s?\d{3}[-.\s]?\d{2}[-.\s]?\d{2}`),
		regexp.MustCompile(`8\s?\(?\d{3}\)?\s?\d{3}[-.\s]?\d{2}[-.\s]?\d{2}`),
		regexp.MustCompile(`\+?\d{1,3}[-.\s]?\(?\d{1,4}\)?[-.\s]?\d{1,4}[-.\s]?\d{1,4}[-.\s]?\d{1,9}`),
		regexp.MustCompile(`(?i)(phone|телефон|тел\.?)\s*[:=]\s*["']?([+\d\s\-\(\)]{7,})["']?`),
	}

	for _, pattern := range patterns {
		text = pattern.ReplaceAllString(text, `[FILTERED_PHONE]`)
	}

	return text
}
