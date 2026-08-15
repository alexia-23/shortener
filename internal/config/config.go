package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
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

	flag.Parse()

	if serverAddress, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddress = serverAddress
	}
	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = baseURL
	}
	return cfg
}
