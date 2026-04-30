package organization_role_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	role_service "github.com/yca-software/2chi-kit/go-api/internals/services/role"
	yca_error "github.com/yca-software/go-common/error"
)

type Handler struct {
	services *services.Services
}

func New(services *services.Services) *Handler {
	return &Handler{services: services}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	group := router.Group("/role")

	group.POST("", h.CreateRole)
	group.GET("", h.ListRoles)
	group.PATCH("/:roleId", h.UpdateRole)
	group.DELETE("/:roleId", h.DeleteRole)
}

// DeleteRole godoc
// @Summary      Delete role
// @Description  Permanently deletes a role by its ID
// @Tags         Organization Role
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        roleId   path      string                         true  "Role ID"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      409      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/role/{roleId} [delete]
func (h *Handler) DeleteRole(c echo.Context) error {
	orgId := c.Param("orgId")
	roleId := c.Param("roleId")

	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	if err := h.services.Role.Delete(&role_service.DeleteRequest{
		OrganizationID: orgId,
		RoleID:         roleId,
	}, accessInfo); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// CreateRole godoc
// @Summary      Create role
// @Description  Creates a new role for an organization
// @Tags         Organization Role
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        role body      role_service.CreateRequest  true  "Role request"
// @Success      201      {object}  models.Role
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      402      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/role [post]
func (h *Handler) CreateRole(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgId := c.Param("orgId")

	var req role_service.CreateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgId

	resp, err := h.services.Role.Create(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, resp)
}

// ListRoles godoc
// @Summary      List roles
// @Description  Retrieves a list of roles for a specific organization
// @Tags         Organization Role
// @Accept       json
// @Produce      json
// @Param        orgId     path      string                         true   "Organization ID"
// @Success      200      {array}   models.Role "List of roles"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      402      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/role [get]
func (h *Handler) ListRoles(c echo.Context) error {
	orgId := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	req := role_service.ListRequest{
		OrganizationID: orgId,
	}

	resp, err := h.services.Role.List(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateRole godoc
// @Summary      Update role
// @Description  Updates a role for an organization
// @Tags         Organization Role
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        roleId   path      string                         true  "Role ID"
// @Param        role body      role_service.UpdateRequest  true  "Role request"
// @Success      200      {object}  models.Role
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      402      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/role/{roleId} [patch]
func (h *Handler) UpdateRole(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgId := c.Param("orgId")
	roleId := c.Param("roleId")

	var req role_service.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgId
	req.RoleID = roleId

	resp, err := h.services.Role.Update(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}
