package main

import (
	"log"
	"net/http"

	"github.com/alexia-23/shortener/internal/config"
	"github.com/alexia-23/shortener/internal/handlers"
	"github.com/alexia-23/shortener/internal/repository"
)

func main() {
	cfg := config.NewConfig()

	shortLinkRepository := repository.NewShortLinkRepository()
	router := handlers.NewRouter(shortLinkRepository, cfg.BaseURL)

	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		log.Fatal(err)
	}
}
