package organization_api_key_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	api_key_service "github.com/yca-software/2chi-kit/go-api/internals/services/api_key"
	yca_error "github.com/yca-software/go-common/error"
)

type Handler struct {
	services *services.Services
}

func New(services *services.Services) *Handler {
	return &Handler{services: services}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	group := router.Group("/api-key")

	group.POST("", h.CreateApiKey)
	group.GET("", h.ListApiKeys)
	group.PATCH("/:apiKeyId", h.UpdateApiKey)
	group.DELETE("/:apiKeyId", h.DeleteApiKey)
}

// CreateApiKey godoc
// @Summary      Create API key
// @Description  Creates a new API key for an organization
// @Tags         Organization API Key
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        apiKey body      api_key_service.CreateRequest  true  "API key request"
// @Success      201      {object}  api_key_service.CreateResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      402      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/api-key [post]
func (h *Handler) CreateApiKey(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")

	var req api_key_service.CreateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgID

	resp, err := h.services.ApiKey.Create(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, resp)
}

// ListApiKeys godoc
// @Summary      List API keys
// @Description  Retrieves a list of API keys for a specific organization
// @Tags         Organization API Key
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Success      200      {array}   models.APIKey "List of API keys"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      402      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/api-key [get]
func (h *Handler) ListApiKeys(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")

	resp, err := h.services.ApiKey.List(&api_key_service.ListRequest{
		OrganizationID: orgID,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateApiKey godoc
// @Summary      Update API key
// @Description  Updates an API key for an organization
// @Tags         Organization API Key
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        apiKeyId   path      string                         true  "API Key ID"
// @Param        apiKey body      api_key_service.UpdateRequest  true  "API key request"
// @Success      200      {object}  models.APIKey
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      402      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/api-key/{apiKeyId} [patch]
func (h *Handler) UpdateApiKey(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")
	apiKeyID := c.Param("apiKeyId")

	var req api_key_service.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgID
	req.ApiKeyID = apiKeyID

	apiKey, err := h.services.ApiKey.Update(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, apiKey)
}

// DeleteApiKey godoc
// @Summary      Delete API key
// @Description  Deletes an API key for an organization
// @Tags         Organization API Key
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        apiKeyId   path      string                         true  "API Key ID"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/api-key/{apiKeyId} [delete]
func (h *Handler) DeleteApiKey(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")
	apiKeyID := c.Param("apiKeyId")

	err := h.services.ApiKey.Delete(&api_key_service.DeleteRequest{
		OrganizationID: orgID,
		ApiKeyID:       apiKeyID,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
