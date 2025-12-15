package config

import (
	"context"
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppName  string
	Env      string
	HTTPPort string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	FirebaseProjectID string
}

func MustLoad() Config {
	cfg := Config{
		AppName:           getenv("APP_NAME", "kinesio-app"),
		Env:               getenv("ENV", "local"),
		HTTPPort:          getenv("HTTP_PORT", "8080"),
		DBHost:            getenv("DB_HOST", "localhost"),
		DBPort:            getenv("DB_PORT", "5432"),
		DBName:            getenv("DB_NAME", "kinesio"),
		DBUser:            getenv("DB_USER", "kinesio"),
		DBPassword:        getenv("DB_PASSWORD", "kinesio"),
		DBSSLMode:         getenv("DB_SSLMODE", "disable"),
		FirebaseProjectID: getenv("FIREBASE_PROJECT_ID", ""),
	}

	// Validaciones mínimas
	if cfg.HTTPPort == "" {
		panic("HTTP_PORT is required")
	}
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.DBUser == "" {
		panic("DB_* configuration is required")
	}
	return cfg
}

func (c Config) PostgresDSN() string {
	// Ej: host=localhost user=foo password=bar dbname=baz port=5432 sslmode=disable TimeZone=UTC
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

func ShutdownContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
