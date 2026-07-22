package main

import (
	"log"
	"net/http"

	"github.com/alexia-23/shortener/internal/handlers"
	"github.com/alexia-23/shortener/internal/repository"
)

func main() {
	shortLinkRepository := repository.NewShortLinkRepository()
	router := handlers.NewRouter(shortLinkRepository)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}
