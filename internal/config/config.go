package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Cfg struct {
	Database   Database
	Logger     Logger
	OpenAI     OpenAI
	Browser    Browser
	Migrations Migrations
}

type Database struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

type Migrations struct {
	Path string
}

type Logger struct {
	Env   string
	Level string
}

type OpenAI struct {
	KeyAI     string
	Model     string
	MaxTokens int
}

type Browser struct {
	Display      string
	Headless     bool
	UserDataDir  string
	BrowsersPath string
}

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
			Path: env("MIGRATIONS_PATH", "file://migrations"),
		},
	}

	return cfg, nil
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
