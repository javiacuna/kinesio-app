package db

import (
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/javiacuna/kinesio-backend/internal/config"
)

func OpenPostgres(cfg config.Config) (*gorm.DB, error) {
	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.Env == "local" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	log.Info().Msg("connected to postgres")
	return db, nil
}
