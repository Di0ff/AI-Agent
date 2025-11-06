package sanitizer

import "regexp"

type TokenSanitizer struct{}

func (s *TokenSanitizer) Sanitize(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(token|токен)\s*[:=]\s*["']?([a-zA-Z0-9_-]{20,})["']?`),
		regexp.MustCompile(`(?i)(api[_-]?key|api[_-]?token)\s*[:=]\s*["']?([a-zA-Z0-9_-]{20,})["']?`),
		regexp.MustCompile(`(?i)(bearer\s+)([a-zA-Z0-9_-]{20,})`),
		regexp.MustCompile(`(?i)(authorization\s*[:=]\s*["']?bearer\s+)([a-zA-Z0-9_-]{20,})["']?`),
		regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`),
		regexp.MustCompile(`pk_[a-zA-Z0-9]{32,}`),
	}

	for _, pattern := range patterns {
		text = pattern.ReplaceAllString(text, `${1}[FILTERED]`)
	}

	return text
}
