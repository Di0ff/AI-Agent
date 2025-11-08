package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"aiAgent/internal/agent"
	"aiAgent/internal/browser"
	"aiAgent/internal/cli"
	"aiAgent/internal/config"
	"aiAgent/internal/database"
	"aiAgent/internal/llm"
	"aiAgent/internal/logger"
	"aiAgent/internal/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m❌ Критическая ошибка:\033[0m %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	log, err := logger.New(cfg.Logger.Env, cfg.Logger.Level)
	if err != nil {
		return fmt.Errorf("ошибка инициализации логгера: %w", err)
	}
	defer log.Sync()

	if err := migrations.Run(cfg, log); err != nil {
		return fmt.Errorf("ошибка выполнения миграций: %w", err)
	}

	db, err := database.New(cfg, log)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	defer db.Close(log)

	repo := database.NewTaskRepository(db.DB)

	var llmClient llm.LLMClient
	if cfg.OpenAI.KeyAI != "" {
		llmClient = llm.NewClient(cfg.OpenAI.KeyAI, cfg.OpenAI.Model, repo)
	}

	br := browser.New(browser.Config{
		Headless:     cfg.Browser.Headless,
		UserDataDir:  cfg.Browser.UserDataDir,
		BrowsersPath: cfg.Browser.BrowsersPath,
		Display:      cfg.Browser.Display,
	})

	// Создаём user input provider для агента
	userInput := cli.NewUserInputProvider()

	// Создаём агента с дефолтными настройками
	ag := agent.New(br, llmClient, repo, log, agent.Config{
		MaxSteps:          50,
		MaxTokens:         cfg.OpenAI.MaxTokens,
		Retries:           3,
		UserInputProvider: userInput,
		UseSubAgents:      true,
		UseMultiStep:      false, // ОТКЛЮЧЕНО: теперь используется новый Reasoning Layer (ReAct pattern)
		MultiStepSize:     5,
		UseMemory:         true, // Включаем Memory для reasoning patterns
	})

	// Создаём context с поддержкой cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Получен сигнал завершения, закрываем приложение...")
		cancel()
	}()

	console := cli.New(repo, log, llmClient, br, ag)
	console.Run(ctx)

	log.Info("Приложение корректно завершено")
	return nil
}
