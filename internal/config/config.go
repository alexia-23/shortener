package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
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
		"short-url-storage.json",
		"path to file storage",
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

	return cfg
}
