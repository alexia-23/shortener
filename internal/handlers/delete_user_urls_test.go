package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexia-23/shortener/internal/auth"
	"github.com/alexia-23/shortener/internal/service"
	"github.com/stretchr/testify/assert"
)

type deleteUserURLsServiceStub struct {
	userID string
	ids    []string
	called bool
}

func (stub *deleteUserURLsServiceStub) CreateShortLink(
	_ context.Context,
	_ string,
) (string, error) {
	return "", nil
}

func (stub *deleteUserURLsServiceStub) CreateShortLinksBatch(
	_ context.Context,
	_ []string,
) ([]string, error) {
	return nil, nil
}

func (stub *deleteUserURLsServiceStub) GetSourceLink(
	_ context.Context,
	_ string,
) (string, bool, error) {
	return "", false, nil
}

func (stub *deleteUserURLsServiceStub) DeleteUserLinks(
	userID string,
	ids []string,
) {
	stub.userID = userID
	stub.ids = append([]string(nil), ids...)
	stub.called = true
}

func TestHandleDeleteUserURLsAccepted(t *testing.T) {
	serviceStub := &deleteUserURLsServiceStub{}

	handler := NewHandler(
		serviceStub,
		"http://localhost:8080",
	)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/user/urls",
		strings.NewReader(
			`["first","second"]`,
		),
	)

	request = request.WithContext(
		auth.ContextWithUserID(
			request.Context(),
			"user-1",
		),
	)

	recorder := httptest.NewRecorder()

	handler.handleDeleteUserURLs(
		recorder,
		request,
	)

	assert.Equal(
		t,
		http.StatusAccepted,
		recorder.Code,
	)

	assert.True(
		t,
		serviceStub.called,
	)

	assert.Equal(
		t,
		"user-1",
		serviceStub.userID,
	)

	assert.Equal(
		t,
		[]string{"first", "second"},
		serviceStub.ids,
	)
}

func TestHandleDeleteUserURLsMalformedBody(t *testing.T) {
	serviceStub := &deleteUserURLsServiceStub{}

	handler := NewHandler(
		serviceStub,
		"http://localhost:8080",
	)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/user/urls",
		strings.NewReader(
			`{"bad":true}`,
		),
	)

	request = request.WithContext(
		auth.ContextWithUserID(
			request.Context(),
			"user-1",
		),
	)

	recorder := httptest.NewRecorder()

	handler.handleDeleteUserURLs(
		recorder,
		request,
	)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	assert.False(
		t,
		serviceStub.called,
	)
}

func TestHandleDeleteUserURLsEmptyList(t *testing.T) {
	serviceStub := &deleteUserURLsServiceStub{}

	handler := NewHandler(
		serviceStub,
		"http://localhost:8080",
	)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/user/urls",
		strings.NewReader(`[]`),
	)

	request = request.WithContext(
		auth.ContextWithUserID(
			request.Context(),
			"user-1",
		),
	)

	recorder := httptest.NewRecorder()

	handler.handleDeleteUserURLs(
		recorder,
		request,
	)

	assert.Equal(
		t,
		http.StatusAccepted,
		recorder.Code,
	)

	assert.True(
		t,
		serviceStub.called,
	)

	assert.Equal(
		t,
		"user-1",
		serviceStub.userID,
	)

	assert.Empty(
		t,
		serviceStub.ids,
	)
}

func TestHandleDeleteUserURLsUnauthorized(t *testing.T) {
	serviceStub := &deleteUserURLsServiceStub{}

	handler := NewHandler(
		serviceStub,
		"http://localhost:8080",
	)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/user/urls",
		strings.NewReader(
			`["first"]`,
		),
	)

	recorder := httptest.NewRecorder()

	handler.handleDeleteUserURLs(
		recorder,
		request,
	)

	assert.Equal(
		t,
		http.StatusUnauthorized,
		recorder.Code,
	)

	assert.False(
		t,
		serviceStub.called,
	)
}

func (*deleteUserURLsServiceStub) GetUserLinks(
	context.Context,
	string,
) ([]service.ShortLink, error) {
	panic("unexpected GetUserLinks call in a DeleteUserURLs test")
}
