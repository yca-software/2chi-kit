package user_refresh_token_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	user_refresh_token_service "github.com/yca-software/2chi-kit/go-api/internals/services/user_refresh_token"
	yca_error "github.com/yca-software/go-common/error"
)

type Handler struct {
	services *services.Services
}

func NewHandler(services *services.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) RegisterRoutes(router *echo.Group) {
	group := router.Group("/token")

	group.DELETE("/:tokenId", h.RevokeRefreshToken)
	group.DELETE("", h.RevokeAllRefreshTokens)
	group.GET("", h.ListActiveRefreshTokens)
}

// ListActiveRefreshTokens godoc
// @Summary      List active refresh tokens
// @Description  Lists all active refresh tokens for a user
// @Tags         user, refresh-token
// @Accept       json
// @Produce      json
// @Success      200      {array}   models.UserRefreshToken
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /user/token [get]
func (h *Handler) ListActiveRefreshTokens(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}
	resp, err := h.services.UserRefreshToken.ListActive(&user_refresh_token_service.ListActiveRequest{
		UserID: accessInfo.User.UserID.String(),
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// RevokeAllRefreshTokens godoc
// @Summary      Revoke all refresh tokens
// @Description  Revokes all refresh tokens for a user
// @Tags         user, refresh-token
// @Accept       json
// @Produce      json
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /user/token [delete]
func (h *Handler) RevokeAllRefreshTokens(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}

	var body struct {
		KeepRefreshToken string `json:"keepRefreshToken"`
	}
	_ = c.Bind(&body)

	if err := h.services.UserRefreshToken.RevokeAll(&user_refresh_token_service.RevokeAllRequest{
		UserID:           accessInfo.User.UserID.String(),
		KeepRefreshToken: body.KeepRefreshToken,
	}, accessInfo); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// RevokeRefreshToken godoc
// @Summary      Revoke refresh token
// @Description  Revokes a refresh token for a user
// @Tags         user, refresh-token
// @Accept       json
// @Produce      json
// @Param        tokenId   path      string                         true  "Refresh Token ID"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /user/token/{tokenId} [delete]
func (h *Handler) RevokeRefreshToken(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}

	if err := h.services.UserRefreshToken.Revoke(&user_refresh_token_service.RevokeRequest{
		UserID:         accessInfo.User.UserID.String(),
		RefreshTokenID: c.Param("tokenId"),
	}, accessInfo); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
