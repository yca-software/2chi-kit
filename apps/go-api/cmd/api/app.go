package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/cron"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/jobs"
	"github.com/yca-software/2chi-kit/go-api/internals/metrics"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	http_transport "github.com/yca-software/2chi-kit/go-api/internals/transport/http"
	yca_error "github.com/yca-software/go-common/error"
	yca_http "github.com/yca-software/go-common/http"
	"github.com/yca-software/go-common/localizer"
	"github.com/yca-software/go-common/logger"
	"github.com/yca-software/go-common/validator"
)

type Application struct {
	App         *echo.Echo
	Server      *http.Server
	Logger      logger.Logger
	Translator  localizer.Translator
	Validator   validator.Validator
	DB          *sqlx.DB
	RedisClient *redis.Client
	JobClient   *jobs.Client
	Metrics     *metrics.Metrics
	Repos       *repositories.Repositories
	RootPath    string
	cancelCron  context.CancelFunc
	stopJobs    func()
}

func NewApplication(appLogger logger.Logger) (*Application, error) {
	rootPath, err := helpers.ModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("resolve module root (go.mod directory): %w", err)
	}

	translator := localizer.NewTranslator(
		constants.SUPPORTED_LANGUAGES,
		constants.DEFAULT_LANGUAGE,
		filepath.Join(rootPath, "locales"),
	)

	validator := validator.New()

	postgresClient, err := database.InitPostgreSQLClient(getPostgresConfig(rootPath), appLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	appMetrics, err := metrics.New(os.Getenv("APP_NAME"))
	if err != nil {
		return nil, fmt.Errorf("failed to init metrics: %w", err)
	}

	e := echo.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		lang := yca_http.GetLanguage(c, constants.SUPPORTED_LANGUAGES, constants.DEFAULT_LANGUAGE)

		if apiErr, ok := yca_error.AsError(err); ok {
			if apiErr.Err != nil {
				appLogger.Log(logger.LogData{
					Level:   "error",
					Message: "API error occurred",
					Error:   apiErr.Err,
					Data: map[string]any{
						"status_code": apiErr.StatusCode,
						"error_code":  apiErr.ErrorCode,
						"path":        c.Path(),
						"method":      c.Request().Method,
					},
				})
			}
			translator.TranslateError(lang, apiErr, nil)
			c.JSON(apiErr.StatusCode, apiErr)
			return
		}

		if he, ok := err.(*echo.HTTPError); ok {
			if he.Internal != nil {
				appLogger.Log(logger.LogData{
					Level:   "error",
					Message: "Echo HTTP error occurred",
					Error:   he.Internal,
					Data: map[string]any{
						"status_code": he.Code,
						"message":     he.Message,
						"path":        c.Path(),
						"method":      c.Request().Method,
					},
				})
			}
			c.JSON(he.Code, map[string]any{
				"errorCode": "0001",
				"message":   he.Message,
			})
			return
		}

		appLogger.Log(logger.LogData{
			Level:   "error",
			Message: "Unhandled error occurred",
			Error:   err,
			Data: map[string]any{
				"path":   c.Path(),
				"method": c.Request().Method,
			},
		})

		e := yca_error.NewInternalServerError(err, "", nil)
		translator.TranslateError(lang, e, nil)
		c.JSON(http.StatusInternalServerError, e)
	}
	bodyLimit := os.Getenv("SERVER_BODY_LIMIT")
	corsAllowOrigins := strings.Split(os.Getenv("SERVER_CORS"), ",")
	e.Use(
		appMetrics.EchoMiddleware(nil),
		middleware.Recover(),
		middleware.RequestID(),
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogLatency:   true,
			LogRemoteIP:  true,
			LogMethod:    true,
			LogURIPath:   true,
			LogRequestID: true,
			LogStatus:    true,
			// Do not log body, headers (Authorization, X-API-Key), or query params that may contain PII or secrets.
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				_, err := fmt.Fprintf(os.Stderr, "[%s] %s %d %s %s %s %s\n",
					v.StartTime.Format(time.RFC3339),
					v.RequestID,
					v.Status,
					v.Latency,
					v.Method,
					v.URIPath,
					v.RemoteIP,
				)
				return err
			},
		}),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: corsAllowOrigins,
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "Accept-Language", "X-API-Key"},
		}),
		middleware.BodyLimit(bodyLimit),
		middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "DENY",
			HSTSMaxAge:            31536000,
			ContentSecurityPolicy: "default-src 'self'",
		}),
	)

	repos := repositories.NewRepositories(postgresClient, appMetrics.GetQueryMetricsHook())

	redisClient, err := database.InitRedisClient(os.Getenv("REDIS_DSN"), appLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}

	srvs, err := services.NewServices(repos, validator, translator, appLogger, rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init services: %w", err)
	}

	http_transport.RegisterGateway(e, &http_transport.GatewayConfig{AccessSecret: os.Getenv("TOKEN_ACCESS_SECRET")}, srvs, repos, appMetrics, redisClient, postgresClient, appLogger)

	readTimeout, err := strconv.Atoi(os.Getenv("SERVER_READ_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert SERVER_READ_TIMEOUT to int: %w", err)
	}
	writeTimeout, err := strconv.Atoi(os.Getenv("SERVER_WRITE_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert SERVER_WRITE_TIMEOUT to int: %w", err)
	}
	idleTimeout, err := strconv.Atoi(os.Getenv("SERVER_IDLE_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert SERVER_IDLE_TIMEOUT to int: %w", err)
	}
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		Handler:      e,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
	}

	var jobClient *jobs.Client
	var stopJobs func()
	if os.Getenv("RABBITMQ_URL") != "" {
		jobClient, err = jobs.NewClient(context.Background(), os.Getenv("RABBITMQ_URL"), appLogger, appMetrics.JobConsumerRecorder())
		if err != nil {
			return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
		}
		jobConsumers := jobs.NewConsumers(srvs, jobClient, appLogger)
		stopJobs = jobConsumers.Start(context.Background())
		appLogger.Log(logger.LogData{
			Level:   "info",
			Message: "Job consumers started",
		})
	}

	var cancelCron context.CancelFunc
	if os.Getenv("CRON_ENABLED") == "true" {
		cancelCron, err = cron.Start(jobClient, appLogger)
		if err != nil {
			return nil, fmt.Errorf("failed to start cron: %w", err)
		}
	}

	return &Application{
		Server:      server,
		Logger:      appLogger,
		Translator:  translator,
		Validator:   validator,
		DB:          postgresClient,
		Metrics:     appMetrics,
		Repos:       repos,
		RootPath:    rootPath,
		RedisClient: redisClient,
		JobClient:   jobClient,
		App:         e,
		cancelCron:  cancelCron,
		stopJobs:    stopJobs,
	}, nil
}

func (app *Application) Start() error {
	return app.Server.ListenAndServe()
}

func (app *Application) Shutdown(ctx context.Context) error {
	app.Logger.Log(logger.LogData{
		Level:   "info",
		Message: "Shutting down server...",
	})

	if app.cancelCron != nil {
		app.cancelCron()
	}
	if app.stopJobs != nil {
		app.stopJobs()
	}

	if app.Server != nil {
		if err := app.Server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	if app.RedisClient != nil {
		if err := app.RedisClient.Close(); err != nil {
			return fmt.Errorf("failed to close Redis connection: %w", err)
		}
	}

	if app.JobClient != nil {
		if err := app.JobClient.Close(); err != nil {
			return fmt.Errorf("failed to close RabbitMQ job client: %w", err)
		}
	}

	app.Logger.Log(logger.LogData{
		Level:   "info",
		Message: "Server shutdown complete",
	})
	return nil
}
