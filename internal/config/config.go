// Package config управляет конфигурацией приложения из переменных окружения (.env).
// Включает валидацию всех обязательных параметров при загрузке.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Cfg содержит всю конфигурацию приложения.
type Cfg struct {
	Database   Database   // Конфигурация PostgreSQL
	Logger     Logger     // Конфигурация логирования
	OpenAI     OpenAI     // Конфигурация OpenAI API
	Browser    Browser    // Конфигурация браузера
	Migrations Migrations // Конфигурация миграций БД
}

// Database содержит параметры подключения к PostgreSQL.
type Database struct {
	Host     string // Хост БД (localhost)
	Port     string // Порт БД (5432)
	Name     string // Имя базы данных
	User     string // Пользователь БД
	Password string // Пароль БД
}

// Migrations содержит путь к SQL миграциям.
type Migrations struct {
	Path string // Путь к миграциям (file://internal/migrations/scripts)
}

// Logger содержит конфигурацию логирования.
type Logger struct {
	Env   string // Окружение (dev, prod, test)
	Level string // Уровень логирования (debug, info, warn, error)
}

// OpenAI содержит конфигурацию для OpenAI API.
type OpenAI struct {
	KeyAI     string // API ключ (должен начинаться с sk-)
	Model     string // Модель (gpt-4o)
	MaxTokens int    // Максимальное количество токенов в запросе
}

// Browser содержит конфигурацию браузера Firefox.
type Browser struct {
	Display      string // DISPLAY для Linux (например :0)
	Headless     bool   // Headless режим (без GUI)
	UserDataDir  string // Директория для сохранения сессий
	BrowsersPath string // Путь к браузерам Playwright
}

// Load загружает конфигурацию из файла .env и переменных окружения.
// Автоматически валидирует все обязательные параметры.
// Возвращает ошибку если конфигурация невалидна.
func Load() (*Cfg, error) {
	_ = godotenv.Load()

	cfg := &Cfg{
		Database: Database{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Name:     os.Getenv("DB_NAME"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASS"),
		},
		Logger: Logger{
			Env:   env("ENV", "dev"),
			Level: env("LOG_LEVEL", "info"),
		},
		OpenAI: OpenAI{
			KeyAI:     os.Getenv("OPENAI_API_KEY"),
			Model:     env("OPENAI_MODEL", "gpt-4o"),
			MaxTokens: envInt("OPENAI_MAX_TOKENS", 4000),
		},
		Browser: Browser{
			Display:      env("DISPLAY", ":0"),
			Headless:     envBool("PW_HEADLESS"),
			UserDataDir:  env("PW_USER_DATA_DIR", "./userdata"),
			BrowsersPath: env("PLAYWRIGHT_BROWSERS_PATH", ""),
		},
		Migrations: Migrations{
			Path: env("MIGRATIONS_PATH", "file://internal/migrations/scripts"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("невалидная конфигурация: %w", err)
	}

	return cfg, nil
}

// Validate проверяет корректность конфигурации
func (c *Cfg) Validate() error {
	var errors []string

	// Проверка обязательных полей Database
	if c.Database.Host == "" {
		errors = append(errors, "DB_HOST обязателен")
	}
	if c.Database.Port == "" {
		errors = append(errors, "DB_PORT обязателен")
	} else {
		if port, err := strconv.Atoi(c.Database.Port); err != nil || port < 1 || port > 65535 {
			errors = append(errors, "DB_PORT должен быть числом от 1 до 65535")
		}
	}
	if c.Database.Name == "" {
		errors = append(errors, "DB_NAME обязателен")
	}
	if c.Database.User == "" {
		errors = append(errors, "DB_USER обязателен")
	}
	// Password может быть пустым для локальной разработки

	// Проверка Logger
	validEnvs := map[string]bool{"dev": true, "prod": true, "test": true}
	if !validEnvs[c.Logger.Env] {
		errors = append(errors, "ENV должен быть одним из: dev, prod, test")
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logger.Level] {
		errors = append(errors, "LOG_LEVEL должен быть одним из: debug, info, warn, error")
	}

	// Проверка OpenAI (опционально, но если указан - должен быть валидным)
	if c.OpenAI.KeyAI != "" {
		if !strings.HasPrefix(c.OpenAI.KeyAI, "sk-") {
			errors = append(errors, "OPENAI_API_KEY должен начинаться с 'sk-'")
		}
	}
	if c.OpenAI.MaxTokens < 100 || c.OpenAI.MaxTokens > 128000 {
		errors = append(errors, "OPENAI_MAX_TOKENS должен быть от 100 до 128000")
	}

	// Проверка Browser
	if c.Browser.UserDataDir != "" {
		// Если указан абсолютный путь - проверяем что родительская директория существует
		if filepath.IsAbs(c.Browser.UserDataDir) {
			parentDir := filepath.Dir(c.Browser.UserDataDir)
			if _, err := os.Stat(parentDir); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("родительская директория для PW_USER_DATA_DIR не существует: %s", parentDir))
			}
		}
	}

	// Проверка Migrations
	if c.Migrations.Path == "" {
		errors = append(errors, "MIGRATIONS_PATH обязателен")
	}

	if len(errors) > 0 {
		return fmt.Errorf("ошибки валидации:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

func env(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func envInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultValue
}

func envBool(key string) bool {
	v := strings.ToLower(os.Getenv(key))
	return v == "true" || v == "1" || v == "yes"
}
