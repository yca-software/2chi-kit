package organization_team_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	team_service "github.com/yca-software/2chi-kit/go-api/internals/services/team"
	organization_team_member_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/team/member"
	yca_error "github.com/yca-software/go-common/error"
)

type Handler struct {
	services *services.Services
}

func New(services *services.Services) *Handler {
	return &Handler{services: services}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	memberHandler := organization_team_member_handler.New(h.services)

	group := router.Group("/team")
	group.GET("", h.ListTeams)
	group.POST("", h.CreateTeam)

	detailGroup := group.Group("/:teamId")
	detailGroup.PATCH("", h.UpdateTeam)
	detailGroup.DELETE("", h.DeleteTeam)

	memberHandler.RegisterEndpoints(detailGroup)
}

// DeleteTeam godoc
// @Summary      Delete team
// @Description  Permanently deletes a team by its ID
// @Tags         Organization Team
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        teamId   path      string                         true  "Team ID"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      409      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/team/{teamId} [delete]
func (h *Handler) DeleteTeam(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgId := c.Param("orgId")
	teamId := c.Param("teamId")

	if err := h.services.Team.Delete(&team_service.DeleteRequest{
		OrganizationID: orgId,
		TeamID:         teamId,
	}, accessInfo); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// CreateTeam godoc
// @Summary      Create team
// @Description  Creates a new team
// @Tags         Organization Team
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        team body      team_service.CreateRequest  true  "Team request"
// @Success      201      {object}  models.Team
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      402      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/team [post]
func (h *Handler) CreateTeam(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgId := c.Param("orgId")

	var req team_service.CreateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgId

	resp, err := h.services.Team.Create(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, resp)
}

// ListTeams godoc
// @Summary      List teams
// @Description  Retrieves a list of teams for a specific organization
// @Tags         Organization Team
// @Accept       json
// @Produce      json
// @Param        orgId     path      string                         true   "Organization ID"
// @Success      200      {array}   models.Team "List of teams"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      402      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/team [get]
func (h *Handler) ListTeams(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgId := c.Param("orgId")

	resp, err := h.services.Team.List(&team_service.ListRequest{
		OrganizationID: orgId,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateTeam godoc
// @Summary      Update team
// @Description  Updates a team for an organization
// @Tags         Organization Team
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        teamId   path      string                         true  "Team ID"
// @Param        team body      team_service.UpdateRequest  true  "Team request"
// @Success      200      {object}  models.Team
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      402      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/team/{teamId} [patch]
func (h *Handler) UpdateTeam(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgId := c.Param("orgId")
	teamId := c.Param("teamId")

	var req team_service.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgId
	req.TeamID = teamId

	resp, err := h.services.Team.Update(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}
