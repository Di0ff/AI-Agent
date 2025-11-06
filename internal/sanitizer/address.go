package sanitizer

import "regexp"

type AddressSanitizer struct{}

func (s *AddressSanitizer) Sanitize(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\d{6}\b`),
		regexp.MustCompile(`(?i)(улица|ул\.?|проспект|пр\.?|проезд|пер\.?|переулок|бульвар|б-р|шоссе|ш\.?)\s+[А-Яа-яЁё\w\s]+(?:,\s*(?:д\.?|дом|стр\.?|строение|корп\.?|корпус|кв\.?|квартира)\s*\d+)?`),
		regexp.MustCompile(`(?i)(г\.?|город|пос\.?|поселок|пгт|село|дер\.?|деревня)\s+[А-Яа-яЁё\w\s-]+`),
		regexp.MustCompile(`(?i)(область|обл\.?|край|республика|респ\.?|район|р-н)\s+[А-Яа-яЁё\w\s-]+`),
		regexp.MustCompile(`(?i)(address|адрес|адр\.?)\s*[:=]\s*["']?([^"'\n]{10,})["']?`),
		regexp.MustCompile(`[А-Яа-яЁё]+\s+[А-Яа-яЁё]+\s+обл\.?,\s*[А-Яа-яЁё\s]+`),
		regexp.MustCompile(`[А-Яа-яЁё]+,\s*(?:ул\.?|пр\.?|б-р|ш\.?)\s+[А-Яа-яЁё\s]+,\s*(?:д\.?|дом)\s*\d+`),
	}

	for _, pattern := range patterns {
		text = pattern.ReplaceAllString(text, `[FILTERED_ADDRESS]`)
	}

	return text
}
