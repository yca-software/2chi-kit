package http_transport

import (
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/yca-software/2chi-kit/go-api/internals/metrics"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	admin_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/admin"
	auth_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/auth"
	health_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/health"
	location_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/location"
	organization_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization"
	paddle_webhook_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/paddle_webhook"
	support_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/support"
	user_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/user"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_token "github.com/yca-software/go-common/token"

	yca_log "github.com/yca-software/go-common/logger"
)

type GatewayConfig struct {
	AccessSecret string
}

func RegisterGateway(e *echo.Echo, gatewayCfg *GatewayConfig, srvs *services.Services, repos *repositories.Repositories, metrics *metrics.Metrics, redisClient *redis.Client, db *sqlx.DB, logger yca_log.Logger) {
	mwares := http_middlewares.NewMiddlewares(&http_middlewares.MiddlewaresConfig{AccessSecret: gatewayCfg.AccessSecret}, repos, metrics, yca_token.HashToken)

	rateLimitConfig := &http_middlewares.RateLimitConfig{
		RedisClient: redisClient,
		Metrics:     metrics,
	}

	healthHandler := health_handler.New(db)
	healthHandler.RegisterEndpoints(e)

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	api := e.Group("/api/v1")

	paddleWebhookHandler := paddle_webhook_handler.New(os.Getenv("PADDLE_WEBHOOK_SECRET"), srvs)
	paddleWebhookHandler.RegisterEndpoints(api)

	authHandler := auth_handler.New(srvs, mwares, rateLimitConfig)
	authHandler.RegisterEndpoints(api)

	userHandler := user_handler.New(srvs, mwares, rateLimitConfig)
	userHandler.RegisterEndpoints(api)

	locationHandler := location_handler.New(srvs, mwares, rateLimitConfig)
	locationHandler.RegisterEndpoints(api)

	adminHandler := admin_handler.New(srvs, mwares)
	adminHandler.RegisterEndpoints(api)

	organizationHandler := organization_handler.New(srvs, mwares, rateLimitConfig)
	organizationHandler.RegisterEndpoints(api)

	supportHandler := support_handler.New(srvs, mwares, rateLimitConfig)
	supportHandler.RegisterEndpoints(api)
}
