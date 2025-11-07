package agent

import (
	"net/url"
	"strings"
)

// DomainSecurityLevel определяет уровень безопасности домена
type DomainSecurityLevel int

const (
	DomainSafe     DomainSecurityLevel = iota // Безопасный домен
	DomainCritical                             // Критичный домен (требует ВСЕГДА подтверждение)
	DomainBlocked                              // Заблокированный домен
)

// DomainSecurity содержит информацию о безопасности домена
type DomainSecurity struct {
	Level       DomainSecurityLevel
	Description string
	Reason      string
}

// ВАЖНО: Этот whitelist используется ТОЛЬКО для безопасности,
// НЕ для упрощения задач агента (соответствие ТЗ!)

// criticalDomains - домены, требующие ВСЕГДА подтверждение пользователя
// Это финансовые и критичные сервисы
var criticalDomains = map[string]string{
	// Банки
	"sberbank.ru":        "Банковские операции",
	"alfabank.ru":        "Банковские операции",
	"vtb.ru":             "Банковские операции",
	"tinkoff.ru":         "Банковские операции",
	"bankofamerica.com":  "Banking operations",
	"chase.com":          "Banking operations",
	"wellsfargo.com":     "Banking operations",
	"citibank.com":       "Banking operations",

	// Платежные системы
	"paypal.com":         "Payment processing",
	"stripe.com":         "Payment processing",
	"square.com":         "Payment processing",
	"venmo.com":          "Payment processing",
	"yandex.ru/money":    "Платежная система",
	"qiwi.com":           "Платежная система",

	// Криптовалюты
	"binance.com":        "Cryptocurrency exchange",
	"coinbase.com":       "Cryptocurrency exchange",
	"kraken.com":         "Cryptocurrency exchange",

	// Государственные сервисы
	"gosuslugi.ru":       "Государственные услуги",
	"nalog.gov.ru":       "Налоговая служба",
	"irs.gov":            "Tax services",
}

// blockedPatterns - паттерны опасных/заблокированных URL
// Это админ-панели и системные страницы
var blockedPatterns = []string{
	"/admin",
	"/administrator",
	"/wp-admin",
	"/phpmyadmin",
	"/cpanel",
	"localhost:8080/admin",
	"127.0.0.1",
}

// CheckDomainSecurity проверяет безопасность домена
func CheckDomainSecurity(urlStr string) DomainSecurity {
	// Парсим URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return DomainSecurity{
			Level:       DomainSafe,
			Description: "Не удалось распарсить URL",
			Reason:      "",
		}
	}

	host := strings.ToLower(parsedURL.Host)
	path := strings.ToLower(parsedURL.Path)

	// Удаляем порт из host
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	// Проверка на заблокированные паттерны
	for _, pattern := range blockedPatterns {
		if strings.Contains(path, pattern) || strings.Contains(host, pattern) {
			return DomainSecurity{
				Level:       DomainBlocked,
				Description: "Заблокированный URL (админ-панель или системная страница)",
				Reason:      "Обнаружен паттерн: " + pattern,
			}
		}
	}

	// Проверка на критичные домены
	for domain, description := range criticalDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return DomainSecurity{
				Level:       DomainCritical,
				Description: description,
				Reason:      "Критичный домен, требует подтверждения пользователя",
			}
		}
	}

	// Проверка на localhost и локальные IP
	if strings.Contains(host, "localhost") || strings.HasPrefix(host, "127.") || strings.HasPrefix(host, "192.168.") {
		return DomainSecurity{
			Level:       DomainCritical,
			Description: "Локальный адрес",
			Reason:      "Операции на локальных адресах требуют подтверждения",
		}
	}

	return DomainSecurity{
		Level:       DomainSafe,
		Description: "Обычный домен",
		Reason:      "",
	}
}

// IsDomainCritical проверяет, является ли домен критичным
func IsDomainCritical(urlStr string) bool {
	security := CheckDomainSecurity(urlStr)
	return security.Level == DomainCritical
}

// IsDomainBlocked проверяет, является ли домен заблокированным
func IsDomainBlocked(urlStr string) bool {
	security := CheckDomainSecurity(urlStr)
	return security.Level == DomainBlocked
}
