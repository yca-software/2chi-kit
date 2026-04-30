package admin_user_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	auth_service "github.com/yca-software/2chi-kit/go-api/internals/services/auth"
	user_service "github.com/yca-software/2chi-kit/go-api/internals/services/user"
	yca_http "github.com/yca-software/go-common/http"
)

type Handler struct {
	services *services.Services
}

func New(services *services.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	group := router.Group("/user")

	group.GET("", h.ListUsers)
	group.GET("/:userId", h.GetUser)
	group.DELETE("/:userId", h.DeleteUser)
	group.POST("/:userId/impersonate", h.ImpersonateUser)
}

// GetUser godoc
// @Summary      Get user
// @Description  Get user details by user ID
// @Tags         Admin User
// @Accept       json
// @Produce      json
// @Param        userId   path      string                         true  "User ID"
// @Success      200      {object}  user_service.GetResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/user/{userId} [get]
func (h *Handler) GetUser(c echo.Context) error {
	userID := c.Param("userId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	resp, err := h.services.User.Get(&user_service.GetRequest{
		UserID: userID,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// ListUsers godoc
// @Summary      List users
// @Description  Retrieves a paginated list of users
// @Tags         Admin User
// @Accept       json
// @Produce      json
// @Param        limit    query     int                            false  "Number of items per page (1-100)"
// @Param        offset   query     int                            false  "Page number (0-based)"
// @Param        search   query     string                         false  "Search phrase"
// @Success      200      {object}  user_service.PaginatedListResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/user [get]
func (h *Handler) ListUsers(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	limit, offset := yca_http.ParseLimitOffset(c, 20, 100)
	search := c.QueryParam("search")

	resp, err := h.services.User.List(&user_service.ListRequest{
		SearchPhrase: search,
		Limit:        limit,
		Offset:       offset,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// ImpersonateUser godoc
// @Summary      Impersonate a user
// @Description  Allows an admin to impersonate another user by user ID
// @Tags         Admin User
// @Accept       json
// @Produce      json
// @Param        userId   path      string                                   true  "User ID"
// @Success      200      {object}  auth_service.AuthenticateResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/user/{userId}/impersonate [post]
func (h *Handler) ImpersonateUser(c echo.Context) error {
	userID := c.Param("userId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	resp, err := h.services.Auth.Impersonate(&auth_service.ImpersonateRequest{
		UserID:    userID,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Soft-deletes a user by user ID (admin only)
// @Tags         Admin User
// @Accept       json
// @Produce      json
// @Param        userId   path      string  true  "User ID"
// @Success      204      {object}  nil
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/user/{userId} [delete]
func (h *Handler) DeleteUser(c echo.Context) error {
	userID := c.Param("userId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	if err := h.services.User.Delete(&user_service.DeleteRequest{
		UserID: userID,
	}, accessInfo); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
