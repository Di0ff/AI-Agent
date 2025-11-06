package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Cfg struct {
	Database   Database
	App        App
	Logger     Logger
	OpenAI     OpenAI
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

type App struct {
	Name        string
	Version     string
	Port        string
	Host        string
	StaticDir   string
	FrontendDir string
}

type Logger struct {
	Env   string
	Level string
}

type OpenAI struct {
	KeyAI     string
	Model     string
	MaxTokens string
}

func Load() (*Cfg, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Файл не найден: %v", err)
	}

	cfg := &Cfg{
		Database: Database{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			Name:     getEnv("DB_NAME"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASS"),
		},

		App: App{
			Name:        getEnv("APP_NAME"),
			Version:     getEnv("APP_VERSION"),
			Port:        getEnv("APP_PORT"),
			Host:        getEnv("APP_HOST"),
			StaticDir:   getEnv("STATIC_DIR"),
			FrontendDir: getEnv("FRONTEND_DIR"),
		},

		Logger: Logger{
			Env:   getEnv("ENV"),
			Level: getEnv("LOG_LEVEL"),
		},

		OpenAI: OpenAI{
			KeyAI:     getEnv("OPENAI_API_KEY"),
			Model:     getEnv("OPENAI_MODEL"),
			MaxTokens: getEnv("OPENAI_MAX_TOKENS"),
		},

		Migrations: Migrations{
			Path: getEnv("MIGRATIONS_PATH"),
		},
	}

	return cfg, nil
}

func getEnv(temp string) string {
	if value := os.Getenv(temp); value != "" {
		return value
	}

	return ""
}
