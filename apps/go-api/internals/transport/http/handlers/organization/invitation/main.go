package organization_invitation_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	invitation_service "github.com/yca-software/2chi-kit/go-api/internals/services/invitation"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_error "github.com/yca-software/go-common/error"
	yca_http "github.com/yca-software/go-common/http"
)

type Handler struct {
	services        *services.Services
	rateLimitConfig *http_middlewares.RateLimitConfig
}

func New(services *services.Services, rateLimitConfig *http_middlewares.RateLimitConfig) *Handler {
	return &Handler{services: services, rateLimitConfig: rateLimitConfig}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	inviteRL := http_middlewares.OrganizationInvitationCreateRateLimitMiddleware(h.rateLimitConfig)
	group := router.Group("/invitation")

	group.POST("", h.CreateInvitation, inviteRL)
	group.GET("", h.ListInvitations)
	group.DELETE("/:invitationId", h.RevokeInvitation)
}

// CreateInvitation godoc
// @Summary      Create invitation
// @Description  Creates a new invitation for a user to join an organization
// @Tags         Organization Invitation
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        invitation body      invitation_service.CreateRequest  true  "Invitation request"
// @Success      201      {object}  invitation_service.CreateResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      409      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      429      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/invitation [post]
func (h *Handler) CreateInvitation(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	var req invitation_service.CreateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}

	if accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}

	req.OrganizationID = orgID
	req.InvitedByID = accessInfo.User.UserID.String()
	req.InvitedByEmail = accessInfo.User.Email
	req.Language = yca_http.GetLanguage(c, constants.SUPPORTED_LANGUAGES, constants.DEFAULT_LANGUAGE)

	resp, err := h.services.Invitation.Create(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, resp)
}

// ListInvitations godoc
// @Summary      List invitations
// @Description  Retrieves a list of invitations for a specific organization
// @Tags         Organization Invitation
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Success      200      {array}   models.Invitation
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/invitation [get]
func (h *Handler) ListInvitations(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	resp, err := h.services.Invitation.List(&invitation_service.ListRequest{
		OrganizationID: orgID,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// RevokeInvitation godoc
// @Summary      Revoke invitation
// @Description  Revokes an invitation for a user to join an organization
// @Tags         Organization Invitation
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        invitationId   path      string                         true  "Invitation ID"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/invitation/{invitationId} [delete]
func (h *Handler) RevokeInvitation(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")
	invitationID := c.Param("invitationId")

	if err := h.services.Invitation.Revoke(&invitation_service.RevokeRequest{
		OrganizationID: orgID,
		InvitationID:   invitationID,
	}, accessInfo); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
