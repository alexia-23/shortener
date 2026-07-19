package main

import (
	"log"
	"net/http"

	"github.com/alexia-23/shortener/internal/handlers"
	"github.com/alexia-23/shortener/internal/repository"
)

func main() {
	shortLinkRepository := repository.NewShortLinkRepository()
	handler := handlers.NewHandler(shortLinkRepository)

	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
