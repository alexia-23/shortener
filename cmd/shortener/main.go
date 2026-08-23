package main

import (
	"net/http"

	"github.com/alexia-23/shortener/internal/config"
	"github.com/alexia-23/shortener/internal/handlers"
	"github.com/alexia-23/shortener/internal/middleware"
	"github.com/alexia-23/shortener/internal/repository"
	"github.com/alexia-23/shortener/internal/service"
	"go.uber.org/zap"
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

	if cfg.FileStoragePath == "" {
		shortLinkRepository = repository.NewMemoryRepository()
	} else {
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
	}

	shortLinkService := service.NewShortLinkService(shortLinkRepository)

	router := handlers.NewRouter(
		shortLinkService,
		cfg.BaseURL,
		middleware.WithLogging(sugar),
		middleware.WithGzip(sugar),
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
