package main

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

var links = make(map[string]string)
var counter int

func main() {
	http.ListenAndServe(":8080", http.HandlerFunc(handler))

}
func handler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		id := strings.TrimPrefix(r.URL.Path, "/")

		originalURL, ok := links[id]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	counter++
	id := strconv.Itoa(counter)
	links[id] = string(body)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + id))
}
