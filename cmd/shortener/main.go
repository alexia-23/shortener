package main

import (
	"log"
	"net/http"

	"github.com/alexia-23/shortener/internal/config"
	"github.com/alexia-23/shortener/internal/handlers"
	"github.com/alexia-23/shortener/internal/repository"
	"github.com/alexia-23/shortener/internal/service"
)

func main() {
	cfg := config.NewConfig()

	shortLinkRepository := repository.NewShortLinkRepository()
	shortLinkService := service.NewShortLinkService(shortLinkRepository)

	router := handlers.NewRouter(
		shortLinkService,
		cfg.BaseURL,
	)

	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		log.Fatal(err)
	}
}
