package main

import (
	"database/sql"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/alexia-23/shortener/internal/config"
	database "github.com/alexia-23/shortener/internal/db"
	"github.com/alexia-23/shortener/internal/handlers"
	"github.com/alexia-23/shortener/internal/middleware"
	"github.com/alexia-23/shortener/internal/repository"
	"github.com/alexia-23/shortener/internal/service"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fallbackLogger := zap.NewExample()
		fallbackLogger.Fatal(
			"failed to initialize logger",
			zap.Error(err),
		)
		return
	}

	defer func() {
		_ = logger.Sync()
	}()

	sugar := logger.Sugar()
	cfg := config.NewConfig()

	var shortLinkRepository service.ShortLinkRepository
	var pinger handlers.Pinger

	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			sugar.Fatalw(
				"failed to initialize database",
				"error", err,
			)
			return
		}

		defer func() {
			if err := db.Close(); err != nil {
				sugar.Errorw(
					"failed to close database",
					"error", err,
				)
			}
		}()

		postgresRepository, err := database.NewPostgresRepository(db)
		if err != nil {
			sugar.Fatalw(
				"failed to initialize postgres repository",
				"error", err,
			)
			return
		}

		shortLinkRepository = postgresRepository
		pinger = postgresRepository
	} else if cfg.FileStoragePath != "" {
		fileRepository, err := repository.NewFileRepository(
			cfg.FileStoragePath,
		)
		if err != nil {
			sugar.Fatalw(
				"failed to initialize file repository",
				"error", err,
			)
			return
		}

		shortLinkRepository = fileRepository
	} else {
		shortLinkRepository = repository.NewMemoryRepository()
	}

	shortLinkService := service.NewShortLinkService(shortLinkRepository)

	router := handlers.NewRouter(
		shortLinkService,
		cfg.BaseURL,
		middleware.WithLogging(sugar),
		middleware.WithGzip(sugar),
	)

	router.Get(
		"/ping",
		handlers.NewPingHandler(pinger),
	)

	sugar.Infow(
		"starting server",
		"address", cfg.ServerAddress,
	)

	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		sugar.Fatalw(
			"server stopped",
			"error", err,
		)
	}
}
