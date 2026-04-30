package http_middlewares_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_error "github.com/yca-software/go-common/error"
)

// failingStore implements limiter.Store and always returns an error from Get.
type failingStore struct{}

func (f *failingStore) Get(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store unavailable")
}

func (f *failingStore) Peek(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store unavailable")
}

func (f *failingStore) Increment(ctx context.Context, key string, count int64, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store unavailable")
}

func (f *failingStore) Reset(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store unavailable")
}

func TestRateLimitMiddleware_FailClosed_WhenStoreGetFails(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "ok")
	}

	config := &http_middlewares.RateLimitConfig{
		Store: &failingStore{},
	}
	mw := http_middlewares.RateLimitMiddleware("5-M", config)
	handler := mw(next)

	err := handler(c)
	require.Error(t, err)

	// Handler must not have been called (fail closed)
	assert.False(t, nextCalled, "next handler should not be called when store fails")

	// Should return 429 (Too Many Requests) via pkg/error
	if appErr, ok := err.(*yca_error.Error); ok {
		assert.Equal(t, http.StatusTooManyRequests, appErr.StatusCode)
	}
}

func TestRateLimitMiddleware_NoConfig_NoOp(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "ok")
	}

	mw := http_middlewares.RateLimitMiddleware("5-M", nil)
	handler := mw(next)

	err := handler(c)
	require.NoError(t, err)
	assert.True(t, nextCalled)
}
