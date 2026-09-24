package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	ServerAddress       string
	BaseURL             string
	FileStoragePath     string
	DatabaseDSN         string
	AuthSecretKey       string
	DeleteBatchSize     int
	DeleteFlushInterval time.Duration
	DeleteWorkers       int
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(
		&cfg.ServerAddress,
		"a",
		"localhost:8080",
		"address to run HTTP server",
	)

	flag.StringVar(
		&cfg.BaseURL,
		"b",
		"http://localhost:8080",
		"base address of the shortened URL",
	)

	flag.StringVar(
		&cfg.FileStoragePath,
		"f",
		"",
		"path to file storage",
	)

	flag.StringVar(
		&cfg.DatabaseDSN,
		"d",
		"",
		"database connection string",
	)

	flag.StringVar(
		&cfg.AuthSecretKey,
		"k",
		"",
		"authentication secret key",
	)

	flag.IntVar(
		&cfg.DeleteBatchSize,
		"delete-batch-size",
		0,
		"maximum deletion batch size (0 uses service default)",
	)

	flag.DurationVar(
		&cfg.DeleteFlushInterval,
		"delete-flush-interval",
		0,
		"deletion batch flush interval, e.g. 500ms or 1s (0 uses service default)",
	)

	flag.IntVar(
		&cfg.DeleteWorkers,
		"delete-workers",
		0,
		"maximum parallel deletion stream writers (0 uses service default)",
	)

	flag.Parse()

	if serverAddress, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddress = serverAddress
	}

	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = baseURL
	}

	if fileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = fileStoragePath
	}

	if databaseDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DatabaseDSN = databaseDSN
	}

	if authSecretKey, ok := os.LookupEnv("AUTH_SECRET_KEY"); ok {
		cfg.AuthSecretKey = authSecretKey
	}

	return cfg
}
