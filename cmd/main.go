package main

import (
	"context"

	"aiAgent/internal/cli"
	"aiAgent/internal/config"
	"aiAgent/internal/database"
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
	console := cli.New(repo, log)
	console.Run(context.Background())
}
