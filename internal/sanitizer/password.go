package sanitizer

import "regexp"

type PasswordSanitizer struct{}

func (s *PasswordSanitizer) Sanitize(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|пароль)\s*[:=]\s*["']?([^"'\s]{3,})["']?`),
		regexp.MustCompile(`(?i)(passwd|pwd)\s*[:=]\s*["']?([^"'\s]{3,})["']?`),
		regexp.MustCompile(`(?i)input\[type=["']password["']\][^>]*value=["']([^"']+)["']`),
		regexp.MustCompile(`(?i)<input[^>]*type=["']password["'][^>]*value=["']([^"']+)["']`),
	}

	for _, pattern := range patterns {
		text = pattern.ReplaceAllString(text, `${1}: [FILTERED]`)
	}

	return text
}
