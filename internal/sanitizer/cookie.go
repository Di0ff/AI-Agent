package sanitizer

import "regexp"

type CookieSanitizer struct{}

func (s *CookieSanitizer) Sanitize(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(cookie|куки)\s*[:=]\s*["']?([^"'\n]{10,})["']?`),
		regexp.MustCompile(`(?i)(session[_-]?id|session[_-]?token)\s*[:=]\s*["']?([a-zA-Z0-9_-]{10,})["']?`),
		regexp.MustCompile(`(?i)(set-cookie\s*[:=]\s*["']?)([^"'\n]{10,})["']?`),
	}

	for _, pattern := range patterns {
		text = pattern.ReplaceAllString(text, `${1}[FILTERED]`)
	}

	return text
}
