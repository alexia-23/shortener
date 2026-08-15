package main

import (
	"log"
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
		panic(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	cfg := config.NewConfig()

	shortLinkRepository := repository.NewShortLinkRepository()
	shortLinkService := service.NewShortLinkService(shortLinkRepository)

	router := handlers.NewRouter(
		shortLinkService,
		cfg.BaseURL,
	)

	loggedRouter := middleware.WithLogging(router, sugar)

	err = http.ListenAndServe(cfg.ServerAddress, loggedRouter)
	if err != nil {
		log.Fatal(err)
	}
}
