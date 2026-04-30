package http_middlewares

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/metrics"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type RateLimitConfig struct {
	RedisClient *redisclient.Client

	// Store is an optional limiter store used only in tests.
	// If set, RedisClient is ignored and no Redis connection is required.
	Store limiter.Store

	// Metrics records rate limit hits when set (optional).
	Metrics *metrics.Metrics
}

func realIPKey(c echo.Context) string {
	key := c.RealIP()
	if key == "" {
		key = c.Request().RemoteAddr
	}
	return key
}

// authenticatedPrincipalKey identifies the caller after PrivateRoute (user id, API key id, or IP).
func authenticatedPrincipalKey(c echo.Context) string {
	v := c.Get("accessInfo")
	if v == nil {
		return "ip:" + realIPKey(c)
	}
	ai, ok := v.(*models.AccessInfo)
	if !ok || ai == nil {
		return "ip:" + realIPKey(c)
	}
	if ai.User != nil {
		return "u:" + ai.User.UserID.String()
	}
	if ai.ApiKey != nil {
		return "k:" + ai.ApiKey.ID.String()
	}
	return "ip:" + realIPKey(c)
}

func emailSendingPrincipalScopedKey(c echo.Context, scope string) string {
	return scope + ":" + authenticatedPrincipalKey(c)
}

func rateLimitMiddlewareWithKey(rate string, config *RateLimitConfig, keyFor func(echo.Context) string) echo.MiddlewareFunc {
	// If no config or no Redis client, return no-op middleware (rate limiting disabled)
	if config == nil || (config.RedisClient == nil && config.Store == nil) {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	var store limiter.Store
	var err error

	if config.Store != nil {
		// Test-only path: use injected store (e.g. in-memory).
		store = config.Store
	} else {
		// Production path: Redis client must be configured.
		if config.RedisClient == nil {
			// Return no-op middleware if Redis is not available
			return func(next echo.HandlerFunc) echo.HandlerFunc {
				return next
			}
		}

		store, err = redisstore.NewStoreWithOptions(config.RedisClient, limiter.StoreOptions{
			Prefix: "limiter",
		})
		if err != nil {
			panic(fmt.Sprintf("failed to create Redis store for rate limiting: %v", err))
		}

		if store == nil {
			panic("failed to initialize rate limiter store: store is nil")
		}
	}

	// Parse rate limit (e.g., "5-M" for 5 requests per minute, "3-H" for 3 per hour)
	rateLimit, err := limiter.NewRateFromFormatted(rate)
	if err != nil {
		panic(fmt.Sprintf("failed to parse rate limit: %v", err))
	}

	instance := limiter.New(store, rateLimit)

	// Ensure instance is not nil
	if instance == nil {
		panic("failed to create rate limiter instance")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := keyFor(c)

			ctx := context.Background()
			limiterCtx, err := instance.Get(ctx, key)
			if err != nil {
				// Fail closed: reject request when rate limiter backend is unavailable
				// to avoid bypassing abuse protection (e.g. Redis down).
				return yca_error.NewTooManyRequestsError(err, constants.TOO_MANY_REQUESTS_CODE, nil)
			}

			// Set rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiterCtx.Limit))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limiterCtx.Remaining))
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", limiterCtx.Reset))

			if limiterCtx.Reached {
				if config.Metrics != nil {
					config.Metrics.RecordRateLimitHit(c.Path())
				}
				rateLimitErr := yca_error.NewTooManyRequestsError(nil, constants.TOO_MANY_REQUESTS_CODE, nil)
				return rateLimitErr
			}

			return next(c)
		}
	}
}

func RateLimitMiddleware(rate string, config *RateLimitConfig) echo.MiddlewareFunc {
	return rateLimitMiddlewareWithKey(rate, config, realIPKey)
}

// AuthRateLimitMiddleware creates a stricter rate limiter for authentication endpoints
func AuthRateLimitMiddleware(config *RateLimitConfig) echo.MiddlewareFunc {
	return RateLimitMiddleware("6-M", config)
}

// PasswordResetRateLimitMiddleware creates a rate limiter for password reset endpoints
func PasswordResetRateLimitMiddleware(config *RateLimitConfig) echo.MiddlewareFunc {
	return RateLimitMiddleware("6-H", config)
}

// LocationAutocompleteRateLimitMiddleware creates a rate limiter for location autocomplete endpoints
func LocationAutocompleteRateLimitMiddleware(config *RateLimitConfig) echo.MiddlewareFunc {
	return RateLimitMiddleware("20-M", config)
}

// SupportSubmitRateLimitMiddleware limits support form submissions that send email to the inbox (per user/API key).
func SupportSubmitRateLimitMiddleware(config *RateLimitConfig) echo.MiddlewareFunc {
	return rateLimitMiddlewareWithKey("6-H", config, func(c echo.Context) string {
		return emailSendingPrincipalScopedKey(c, "support_submit")
	})
}

// ResendVerificationEmailRateLimitMiddleware limits verification email resends (per user/API key).
func ResendVerificationEmailRateLimitMiddleware(config *RateLimitConfig) echo.MiddlewareFunc {
	return rateLimitMiddlewareWithKey("6-H", config, func(c echo.Context) string {
		return emailSendingPrincipalScopedKey(c, "resend_verification")
	})
}

// OrganizationInvitationCreateRateLimitMiddleware limits invitation emails (per user/API key).
func OrganizationInvitationCreateRateLimitMiddleware(config *RateLimitConfig) echo.MiddlewareFunc {
	return rateLimitMiddlewareWithKey("100-D", config, func(c echo.Context) string {
		return emailSendingPrincipalScopedKey(c, "org_invitation_create")
	})
}
