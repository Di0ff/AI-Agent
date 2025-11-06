package database

import (
	"fmt"
	"time"

	"aiAgent/internal/config"
	"aiAgent/internal/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func New(cfg *config.Cfg, log *logger.Zap) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.User,
		cfg.Database.Password,
	)

	log.Info("Подключение к БД",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.Name),
	)

	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		log.Error("Ошибка получения пулла", zap.Error(err))
		return nil, err
	}

	pullDB, err := db.DB()
	if err != nil {
		log.Error("Ошибка получения пулла", zap.Error(err))
		return nil, err
	}

	pullDB.SetMaxIdleConns(10)
	pullDB.SetMaxOpenConns(100)
	pullDB.SetConnMaxLifetime(time.Hour)

	log.Info("Успех! Подключение к БД")

	return &Database{DB: db}, nil
}

func (db *Database) Close(log *logger.Zap) error {
	pullDB, err := db.DB.DB()
	if err != nil {
		log.Error("Ошибка получения пулла", zap.Error(err))
		return err
	}

	return pullDB.Close()
}
