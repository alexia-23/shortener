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

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
