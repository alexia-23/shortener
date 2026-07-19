package handlers

import (
	"io"
	"net/http"
	"strconv"
)

func handleCreateShortLink(w http.ResponseWriter, r *http.Request) {
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

	mutex.Lock()
	counter++
	id := strconv.Itoa(counter)
	links[id] = string(body)
	mutex.Unlock()

	shortURL := "http://localhost:8080/" + id

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		return
	}
}
