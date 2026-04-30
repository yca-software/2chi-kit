package support_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	support_service "github.com/yca-software/2chi-kit/go-api/internals/services/support"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
)

type Handler struct {
	services        *services.Services
	middlewares     http_middlewares.Middlewares
	rateLimitConfig *http_middlewares.RateLimitConfig
}

func New(srvs *services.Services, mwares http_middlewares.Middlewares, rateLimitConfig *http_middlewares.RateLimitConfig) *Handler {
	return &Handler{services: srvs, middlewares: mwares, rateLimitConfig: rateLimitConfig}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	supportRL := http_middlewares.SupportSubmitRateLimitMiddleware(h.rateLimitConfig)
	group := router.Group("/support", h.middlewares.PrivateRoute)
	group.POST("", h.Submit, supportRL)
}

// Submit godoc
// @Summary      Submit support request
// @Description  Sends a support message to the configured support inbox by email
// @Tags         support
// @Accept       json
// @Produce      json
// @Param        body  body      support_service.SubmitRequest  true  "Support request"
// @Success      204   "No Content"
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Failure      422   {object}  models.ErrorResponse
// @Failure      429   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Router       /support [post]
func (h *Handler) Submit(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	var req support_service.SubmitRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if req.Subject == "" {
		req.Subject = "(no subject)"
	}
	req.UserAgent = c.Request().UserAgent()

	if err := h.services.Support.Submit(&req, accessInfo); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
