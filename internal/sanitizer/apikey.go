package sanitizer

import "regexp"

type APIKeySanitizer struct{}

func (s *APIKeySanitizer) Sanitize(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api[_-]?key|api[_-]?secret)\s*[:=]\s*["']?([a-zA-Z0-9_-]{20,})["']?`),
		regexp.MustCompile(`(?i)(secret[_-]?key|secret[_-]?token)\s*[:=]\s*["']?([a-zA-Z0-9_-]{20,})["']?`),
		regexp.MustCompile(`(?i)(access[_-]?token|access[_-]?key)\s*[:=]\s*["']?([a-zA-Z0-9_-]{20,})["']?`),
	}

	for _, pattern := range patterns {
		text = pattern.ReplaceAllString(text, `${1}: [FILTERED]`)
	}

	return text
}
