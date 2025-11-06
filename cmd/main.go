package main

import (
	"aiAgent/internal/config"
	"aiAgent/internal/database"
	"aiAgent/internal/logger"
	"aiAgent/internal/migrations"
	"aiAgent/internal/server"
	"context"

	"go.uber.org/zap"
)

func main() {
	// 1️⃣ загрузка конфига
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// 2️⃣ логгер
	log, err := logger.New(cfg.Logger.Env, cfg.Logger.Level)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	// 3️⃣ миграции
	if err := migrations.Run(cfg, log); err != nil {
		log.Fatal("Ошибка миграций", zap.Error(err))
	}

	// 4️⃣ подключение к БД
	db, err := database.New(cfg, log)
	if err != nil {
		log.Fatal("Ошибка подключения к БД", zap.Error(err))
	}
	defer db.Close(log)

	// 5️⃣ запуск сервера
	srv := server.New(cfg, log, db)
	if err := srv.Run(context.Background()); err != nil {
		log.Fatal("Ошибка запуска сервера", zap.Error(err))
	}
}
