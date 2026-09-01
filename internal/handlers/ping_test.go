package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type databasePingerMock struct {
	err error
}

func (mock *databasePingerMock) PingContext(ctx context.Context) error {
	return mock.err
}

func TestPingHandler(t *testing.T) {
	tests := []struct {
		name       string
		pingError  error
		wantStatus int
	}{
		{
			name:       "database is available",
			pingError:  nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "database is unavailable",
			pingError:  errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &databasePingerMock{
				err: test.pingError,
			}

			handler := NewPingHandler(db)

			request := httptest.NewRequest(
				http.MethodGet,
				"/ping",
				nil,
			)

			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			assert.Equal(
				t,
				test.wantStatus,
				recorder.Code,
			)
		})
	}
}
