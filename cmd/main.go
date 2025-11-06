package main

import (
	"context"

	"aiAgent/internal/browser"
	"aiAgent/internal/cli"
	"aiAgent/internal/config"
	"aiAgent/internal/database"
	"aiAgent/internal/llm"
	"aiAgent/internal/logger"
	"aiAgent/internal/migrations"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Logger.Env, cfg.Logger.Level)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	if err := migrations.Run(cfg, log); err != nil {
		log.Fatal("Ошибка миграций", zap.Error(err))
	}

	db, err := database.New(cfg, log)
	if err != nil {
		log.Fatal("Ошибка подключения к БД", zap.Error(err))
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

	console := cli.New(repo, log, llmClient, br)
	console.Run(context.Background())
}
