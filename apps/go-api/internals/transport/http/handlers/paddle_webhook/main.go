package paddle_webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	paddle_service "github.com/yca-software/2chi-kit/go-api/internals/services/paddle"
)

const (
	headerPaddleSignature = "Paddle-Signature"
	maxSignatureAge       = 5 * time.Minute
)

// Handler handles Paddle webhook HTTP requests (public, no auth). Signature is verified.
type Handler struct {
	paddle        paddle_service.Service
	webhookSecret string
}

// New creates a Paddle webhook handler.
func New(webhookSecret string, srvs *services.Services) *Handler {
	return &Handler{webhookSecret: webhookSecret, paddle: srvs.Paddle}
}

// RegisterEndpoints registers POST /webhooks/paddle. Must use raw body for signature verification.
func (h *Handler) RegisterEndpoints(router *echo.Group) {
	router.POST("/webhooks/paddle", h.PaddleWebhook)
}

// PaddleWebhook receives Paddle webhook notifications, verifies signature, and delegates to the Paddle service.
func (h *Handler) PaddleWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to read body"})
	}

	signature := strings.TrimSpace(c.Request().Header.Get(headerPaddleSignature))
	if signature == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing Paddle-Signature"})
	}

	if !verifyPaddleSignature(h.webhookSecret, body, signature) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
	}

	if err := h.paddle.HandleWebhook(body, signature); err != nil {
		return err // yca_error types are handled by global error middleware
	}

	return c.JSON(http.StatusOK, map[string]string{"success": "true"})
}

// verifyPaddleSignature checks Paddle-Signature (ts=...;h1=...) and rejects old events.
func verifyPaddleSignature(secret string, body []byte, header string) bool {
	ts, h1, ok := parsePaddleSignature(header)
	if !ok || h1 == "" {
		return false
	}
	timestamp, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}
	eventTime := time.Unix(timestamp, 0)
	if time.Since(eventTime) > maxSignatureAge || eventTime.After(time.Now().Add(time.Minute)) {
		return false
	}
	payload := ts + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(h1))
}

func parsePaddleSignature(header string) (ts, h1 string, ok bool) {
	for _, part := range strings.Split(header, ";") {
		part = strings.TrimSpace(part)
		if i := strings.Index(part, "="); i > 0 {
			key, val := part[:i], part[i+1:]
			switch strings.TrimSpace(key) {
			case "ts":
				ts = strings.TrimSpace(val)
			case "h1":
				h1 = strings.TrimSpace(val)
			}
		}
	}
	return ts, h1, ts != "" && h1 != ""
}
