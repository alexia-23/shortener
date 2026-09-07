package handlers

import (
	"context"
	"net/http"
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

type noopPinger struct{}

func (noopPinger) PingContext(_ context.Context) error {
	return nil
}

func NewPingHandler(pinger Pinger) http.HandlerFunc {
	if pinger == nil {
		pinger = noopPinger{}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := pinger.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
