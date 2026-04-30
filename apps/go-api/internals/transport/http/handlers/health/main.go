package health_handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterEndpoints(e *echo.Echo) {
	e.GET("/health", h.Health)
	e.GET("/ready", h.Ready)
}

// Health godoc
// @Summary      Liveness check
// @Description  Returns 200 if the application is running. Used by orchestrators to determine if the process is alive.
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// Ready godoc
// @Summary      Readiness check
// @Description  Returns 200 if the application can serve traffic (e.g. database is reachable). Used by orchestrators to determine if the instance should receive traffic.
// @Tags         health
// @Produce      json
// @Success      200   {object}  map[string]string
// @Failure      503   {object}  map[string]string
// @Router       /ready [get]
func (h *Handler) Ready(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable",
			"reason": "database not configured",
		})
	}
	if err := h.db.PingContext(c.Request().Context()); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable",
			"reason": "database connection failed",
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
