package migrations

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"go.uber.org/zap"

	"aiAgent/internal/config"
	"aiAgent/internal/logger"
)

func build(cfg *config.Cfg) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
}

func Run(cfg *config.Cfg, log *logger.Zap) error {
	log.Info("Запуск миграций")

	dsn := build(cfg)

	m, err := migrate.New(
		cfg.Migrations.Path,
		dsn,
	)

	if err != nil {
		log.Error("Ошибка инициализации миграций", zap.Error(err))
		return nil
	}

	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			log.Error("Ошибка закрытия миграций",
				zap.Error(srcErr),
				zap.NamedError("db_error", dbErr),
			)
		}
	}()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("Нет изменений (актуал)")
			return nil
		}

		log.Error("Ошибка применения миграций", zap.Error(err))
		return nil
	}

	log.Info("Успех! Миграции применились")
	return nil
}

func Down(cfg *config.Cfg, log *logger.Zap) error {
	log.Warn("Откат миграции")

	dsn := build(cfg)

	m, err := migrate.New(
		cfg.Migrations.Path,
		dsn,
	)

	if err != nil {
		log.Error("Ошибка инициализации миграций", zap.Error(err))
		return nil
	}

	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			log.Error("Ошибка закрытия миграций",
				zap.Error(srcErr),
				zap.NamedError("db_error", dbErr),
			)
		}
	}()

	if err := m.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("Нет миграций для отката")
			return nil
		}
		log.Error("Ошибка отката миграций", zap.Error(err))
		return nil
	}

	log.Info("Успех! Миграция откатилась")
	return nil
}
