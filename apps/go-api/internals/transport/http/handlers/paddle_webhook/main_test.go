package paddle_webhook_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	paddle_service "github.com/yca-software/2chi-kit/go-api/internals/services/paddle"
	paddle_webhook "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/paddle_webhook"
)

func buildSignature(secret string, body []byte, ts int64) string {
	tsStr := strconv.FormatInt(ts, 10)
	payload := tsStr + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return "ts=" + tsStr + ";h1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestPaddleWebhook_MissingSignature(t *testing.T) {
	mockPaddle := paddle_service.NewMockPaddleService()
	srvs := &services.Services{Paddle: mockPaddle}
	h := paddle_webhook.New("test-secret", srvs)
	e := echo.New()
	e.POST("/webhooks/paddle", h.PaddleWebhook)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/paddle", strings.NewReader(`{"event":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing")
}

func TestPaddleWebhook_InvalidSignature(t *testing.T) {
	mockPaddle := paddle_service.NewMockPaddleService()
	srvs := &services.Services{Paddle: mockPaddle}
	h := paddle_webhook.New("test-secret", srvs)
	e := echo.New()
	e.POST("/webhooks/paddle", h.PaddleWebhook)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/paddle", strings.NewReader(`{"event":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Paddle-Signature", "ts=1700000000;h1=invalid")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid signature")
}

func TestPaddleWebhook_Success(t *testing.T) {
	secret := "webhook-secret"
	body := []byte(`{"event":"subscription.created"}`)
	ts := time.Now().Unix()
	sig := buildSignature(secret, body, ts)

	mockPaddle := paddle_service.NewMockPaddleService()
	mockPaddle.On("HandleWebhook", body, sig).Return(nil)
	srvs := &services.Services{Paddle: mockPaddle}
	h := paddle_webhook.New(secret, srvs)
	e := echo.New()
	e.POST("/webhooks/paddle", h.PaddleWebhook)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/paddle", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Paddle-Signature", sig)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "success")
	mockPaddle.AssertExpectations(t)
}
