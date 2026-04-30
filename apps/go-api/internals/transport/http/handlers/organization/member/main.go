package organization_member_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	organization_member_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization_member"
	yca_error "github.com/yca-software/go-common/error"
)

type Handler struct {
	services *services.Services
}

func New(services *services.Services) *Handler {
	return &Handler{services: services}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	group := router.Group("/member")

	group.GET("", h.ListOrganizationMembers)
	group.PATCH("/:memberId/role", h.UpdateOrganizationMemberRole)
	group.DELETE("/:memberId", h.RemoveOrganizationMember)
}

// ListOrganizationMembers godoc
// @Summary      List organization members
// @Description  Retrieves a list of members for a specific organization
// @Tags         Organization Member
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Success      200      {array}   models.OrganizationMemberWithUser
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/member [get]
func (h *Handler) ListOrganizationMembers(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")

	resp, err := h.services.OrganizationMember.ListByOrganization(&organization_member_service.ListByOrganizationRequest{
		OrganizationID: orgID,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// RemoveOrganizationMember godoc
// @Summary      Remove organization member
// @Description  Removes a member from an organization
// @Tags         Organization Member
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        memberId   path      string                         true  "Member ID"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/member/{memberId} [delete]
func (h *Handler) RemoveOrganizationMember(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")
	memberID := c.Param("memberId")

	if err := h.services.OrganizationMember.Remove(&organization_member_service.RemoveRequest{
		OrganizationID: orgID,
		MemberID:       memberID,
	}, accessInfo); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateOrganizationMemberRole godoc
// @Summary      Update organization member role
// @Description  Updates the role of a member in an organization
// @Tags         Organization Member
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        memberId   path      string                         true  "Member ID"
// @Param        role body      organization_member_service.UpdateRequest  true  "Role request"
// @Success      200      {object}  models.OrganizationMemberWithUser
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/member/{memberId}/role [patch]
func (h *Handler) UpdateOrganizationMemberRole(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")
	memberID := c.Param("memberId")

	var req organization_member_service.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgID
	req.MemberID = memberID

	resp, err := h.services.OrganizationMember.Update(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}
