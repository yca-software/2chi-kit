package organization_team_member_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	team_member_service "github.com/yca-software/2chi-kit/go-api/internals/services/team_member"
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

	group.GET("", h.ListTeamMembers)
	group.POST("", h.AddTeamMember)
	group.DELETE("/:memberId", h.RemoveTeamMember)
}

// AddTeamMember godoc
// @Summary      Add team member
// @Description  Adds a member to a team
// @Tags         Organization Team Member
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        teamId  path      string                         true  "Team ID"
// @Param        member body      team_member_service.AddRequest  true  "Team member request"
// @Success      201      {object}  models.TeamMemberWithUser
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      409      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/team/{teamId}/member [post]
func (h *Handler) AddTeamMember(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgId := c.Param("orgId")
	teamId := c.Param("teamId")

	var req team_member_service.AddRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgId
	req.TeamID = teamId

	resp, err := h.services.TeamMember.Add(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, resp)
}

// ListTeamMembers godoc
// @Summary      List team members
// @Description  Retrieves a list of members for a specific team
// @Tags         Organization Team Member
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        teamId  path      string                         true  "Team ID"
// @Success      200      {array}   models.TeamMemberWithUser
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/team/{teamId}/member [get]
func (h *Handler) ListTeamMembers(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgId := c.Param("orgId")
	teamId := c.Param("teamId")

	resp, err := h.services.TeamMember.ListByTeam(&team_member_service.ListByTeamRequest{
		OrganizationID: orgId,
		TeamID:         teamId,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// RemoveTeamMember godoc
// @Summary      Remove team member
// @Description  Removes a member from a team
// @Tags         Organization Team Member
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        teamId   path      string                         true  "Team ID"
// @Param        memberId   path      string                         true  "Member ID"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/team/{teamId}/member/{memberId} [delete]
func (h *Handler) RemoveTeamMember(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgId := c.Param("orgId")
	teamId := c.Param("teamId")
	memberId := c.Param("memberId")

	if err := h.services.TeamMember.Remove(&team_member_service.RemoveRequest{
		OrganizationID: orgId,
		TeamID:         teamId,
		MemberID:       memberId,
	}, accessInfo); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
