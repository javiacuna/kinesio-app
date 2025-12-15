package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/javiacuna/kinesio-backend/internal/config"
	"github.com/javiacuna/kinesio-backend/internal/db"
	httpapi "github.com/javiacuna/kinesio-backend/internal/http"
)

func main() {
	_ = godotenv.Load()

	cfg := config.MustLoad()

	zerolog.TimeFieldFormat = time.RFC3339
	if cfg.Env == "local" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	log.Info().Str("app", cfg.AppName).Str("env", cfg.Env).Msg("starting")

	gin.SetMode(gin.ReleaseMode)
	if cfg.Env == "local" {
		gin.SetMode(gin.DebugMode)
	}

	gormDB, err := db.OpenPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}

	router := httpapi.NewRouter(cfg, gormDB)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("http server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("http server crashed")
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	log.Info().Msg("shutting down")

	ctx, cancel := config.ShutdownContext()
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
	log.Info().Msg("bye")
}
