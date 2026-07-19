package main

import (
	"log"
	"net/http"

	"github.com/alexia-23/shortener/internal/handlers"
)

func main() {
	err := http.ListenAndServe(
		":8080",
		http.HandlerFunc(handlers.Handler),
	)
	if err != nil {
		log.Fatal(err)
	}
}
