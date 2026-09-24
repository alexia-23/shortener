package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/alexia-23/shortener/internal/auth"
	"github.com/alexia-23/shortener/internal/config"
	database "github.com/alexia-23/shortener/internal/db"
	"github.com/alexia-23/shortener/internal/handlers"
	"github.com/alexia-23/shortener/internal/middleware"
	"github.com/alexia-23/shortener/internal/repository"
	"github.com/alexia-23/shortener/internal/service"
)

const (
	authSecretKeySize = 32
	shutdownTimeout   = 5 * time.Second
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fallbackLogger := zap.NewExample()

		fallbackLogger.Error(
			"failed to initialize logger",
			zap.Error(err),
		)

		_ = fallbackLogger.Sync()
		os.Exit(1)
	}

	sugar := logger.Sugar()

	err = run(sugar)
	if err != nil {
		sugar.Errorw(
			"application stopped",
			"error",
			err,
		)
	}

	_ = logger.Sync()

	if err != nil {
		os.Exit(1)
	}
}

func run(sugar *zap.SugaredLogger) error {
	cfg := config.NewConfig()

	var shortLinkRepository service.URLRepository
	var pinger handlers.Pinger

	if cfg.DatabaseDSN != "" {
		db, err := sql.Open(
			"pgx",
			cfg.DatabaseDSN,
		)
		if err != nil {
			return fmt.Errorf(
				"initialize database: %w",
				err,
			)
		}

		defer func() {
			if err := db.Close(); err != nil {
				sugar.Errorw(
					"failed to close database",
					"error",
					err,
				)
			}
		}()

		postgresRepository, err := database.NewPostgresRepository(db)
		if err != nil {
			return fmt.Errorf(
				"initialize postgres repository: %w",
				err,
			)
		}

		shortLinkRepository = postgresRepository
		pinger = postgresRepository
	} else if cfg.FileStoragePath != "" {
		fileRepository, err := repository.NewFileRepository(
			cfg.FileStoragePath,
		)
		if err != nil {
			return fmt.Errorf(
				"initialize file repository: %w",
				err,
			)
		}

		shortLinkRepository = fileRepository
	} else {
		shortLinkRepository = repository.NewMemoryRepository()
	}

	shortLinkService := service.NewShortLinkService(
		shortLinkRepository,
		service.WithDeleteConfig(
			service.DeleteConfig{
				BatchSize:     cfg.DeleteBatchSize,
				FlushInterval: cfg.DeleteFlushInterval,
				MaxWorkers:    cfg.DeleteWorkers,
			},
		),
	)

	defer shortLinkService.Close()

	authSecretKey := cfg.AuthSecretKey
	if authSecretKey == "" {
		var err error

		authSecretKey, err = generateAuthSecretKey()
		if err != nil {
			return fmt.Errorf(
				"generate authentication secret key: %w",
				err,
			)
		}
	}

	cookieSigner := auth.NewCookieSigner(
		authSecretKey,
	)

	router := handlers.NewRouter(
		shortLinkService,
		cfg.BaseURL,
		middleware.WithLogging(sugar),
		middleware.WithGzip(sugar),
		middleware.Authentication(cookieSigner),
	)

	router.Get(
		"/ping",
		handlers.NewPingHandler(pinger),
	)

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverError := make(
		chan error,
		1,
	)

	sugar.Infow(
		"starting server",
		"address",
		cfg.ServerAddress,
	)

	go func() {
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf(
				"run HTTP server: %w",
				err,
			)
		}

		return nil

	case <-ctx.Done():
		sugar.Infow(
			"shutting down server",
		)

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				return errors.Join(
					fmt.Errorf(
						"shutdown HTTP server: %w",
						err,
					),
					fmt.Errorf(
						"close HTTP server: %w",
						closeErr,
					),
				)
			}

			return fmt.Errorf(
				"shutdown HTTP server: %w",
				err,
			)
		}

		err := <-serverError
		if err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf(
				"run HTTP server: %w",
				err,
			)
		}

		return nil
	}
}

func generateAuthSecretKey() (string, error) {
	randomBytes := make(
		[]byte,
		authSecretKeySize,
	)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf(
			"generate random authentication secret: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(
		randomBytes,
	), nil
}
