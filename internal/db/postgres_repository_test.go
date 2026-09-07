package db

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unique violation",
			err: &pgconn.PgError{
				Code: pgerrcode.UniqueViolation,
			},
			want: true,
		},
		{
			name: "wrapped unique violation",
			err: fmt.Errorf(
				"insert short link: %w",
				&pgconn.PgError{Code: pgerrcode.UniqueViolation},
			),
			want: true,
		},
		{
			name: "other postgres error",
			err:  &pgconn.PgError{Code: "23502"},
			want: false,
		},
		{
			name: "non postgres error",
			err:  errors.New("database error"),
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, isUniqueViolation(test.err))
		})
	}
}
