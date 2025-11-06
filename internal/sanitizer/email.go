package sanitizer

import "regexp"

type EmailSanitizer struct{}

func (s *EmailSanitizer) Sanitize(text string) string {
	emailPattern := regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)
	return emailPattern.ReplaceAllString(text, `[FILTERED_EMAIL]`)
}
