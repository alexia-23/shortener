package handlers

import (
	"context"
	"net/http"
)

type PersistenceService interface {
	PingContext(ctx context.Context) error
}

func NewPingHandler(persistence PersistenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := persistence.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
