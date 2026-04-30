package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/yca-software/go-common/logger"
)

// @title           Example Go API
// @version         1.0
// @description     This is the Example Go API server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      https://api.yca.software // TODO: UPDATE
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for authentication. Can be used as an alternative to JWT Bearer token.
func main() {
	appLogger := logger.New()

	app, err := NewApplication(appLogger)
	if err != nil {
		appLogger.Log(logger.LogData{
			Level:   "fatal",
			Message: "Failed to initialize application",
			Error:   err,
		})
		os.Exit(1)
	}

	serverErr := make(chan error, 1)
	go func() {
		app.Logger.Log(logger.LogData{
			Level:   "info",
			Message: fmt.Sprintf("Starting server on port %s", os.Getenv("SERVER_PORT")),
		})
		if err := app.Start(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		app.Logger.Log(logger.LogData{
			Level:   "info",
			Message: fmt.Sprintf("Received signal: %v", sig),
		})
	case err := <-serverErr:
		app.Logger.Log(logger.LogData{
			Level:   "fatal",
			Message: "Server error",
			Error:   err,
		})
		os.Exit(1)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		app.Logger.Log(logger.LogData{
			Level:   "fatal",
			Message: "Shutdown error",
			Error:   err,
		})
		os.Exit(1)
	}
}
