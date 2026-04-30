package location_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_error "github.com/yca-software/go-common/error"
)

type Handler struct {
	services        *services.Services
	middlewares     http_middlewares.Middlewares
	rateLimitConfig *http_middlewares.RateLimitConfig
}

func New(services *services.Services, middlewares http_middlewares.Middlewares, rateLimitConfig *http_middlewares.RateLimitConfig) *Handler {
	return &Handler{
		services:        services,
		middlewares:     middlewares,
		rateLimitConfig: rateLimitConfig,
	}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	locationAutocompleteRateLimit := http_middlewares.LocationAutocompleteRateLimitMiddleware(h.rateLimitConfig)
	router.GET("/location/autocomplete", h.Autocomplete, locationAutocompleteRateLimit, h.middlewares.PrivateRoute)
}

// Autocomplete godoc
// @Summary      Get location autocomplete suggestions
// @Description  Returns location suggestions based on input query
// @Tags         location
// @Accept       json
// @Produce      json
// @Param        input   query     string  true  "Search input"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /location/autocomplete [get]
func (h *Handler) Autocomplete(c echo.Context) error {
	input := c.QueryParam("input")
	if input == "" {
		return yca_error.NewBadRequestError(nil, "input parameter is required", nil)
	}

	resp, err := h.services.Google.AutocompleteLocation(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}
