package handlers

import (
	"context"
	"net/http"
)

type DatabasePinger interface {
	PingContext(ctx context.Context) error
}

func NewPingHandler(db DatabasePinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
