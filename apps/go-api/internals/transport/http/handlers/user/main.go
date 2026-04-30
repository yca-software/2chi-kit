package user_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	auth_service "github.com/yca-software/2chi-kit/go-api/internals/services/auth"
	user_service "github.com/yca-software/2chi-kit/go-api/internals/services/user"
	user_refresh_token_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/user/refresh_token"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_error "github.com/yca-software/go-common/error"
)

type Handler struct {
	services        *services.Services
	middlewares     http_middlewares.Middlewares
	rateLimitConfig *http_middlewares.RateLimitConfig
}

func New(srvs *services.Services, mwares http_middlewares.Middlewares, rateLimitConfig *http_middlewares.RateLimitConfig) *Handler {
	return &Handler{
		services:        srvs,
		middlewares:     mwares,
		rateLimitConfig: rateLimitConfig,
	}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	resendRL := http_middlewares.ResendVerificationEmailRateLimitMiddleware(h.rateLimitConfig)
	group := router.Group("/user", h.middlewares.PrivateRoute)

	group.PATCH("/language", h.UpdateProfileLanguage)
	group.PATCH("/profile", h.UpdateProfile)
	group.PATCH("/password", h.ChangePassword)
	group.PATCH("/terms", h.AcceptTerms)
	group.POST("/resend-verification-email", h.ResendVerificationEmail, resendRL)
	group.GET("", h.GetCurrentUser)

	refreshTokenHandler := user_refresh_token_handler.NewHandler(h.services)
	refreshTokenHandler.RegisterRoutes(group)
}

// AcceptTerms godoc
// @Summary      Accept terms of service
// @Description  Records acceptance of the current terms of service for the authenticated user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        body   body      user_service.AcceptTermsRequest  true  "Terms acceptance (termsVersion)"
// @Success      200   {object}  models.User
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Failure      422   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /user/terms [patch]
func (h *Handler) AcceptTerms(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}

	var req user_service.AcceptTermsRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	req.UserID = accessInfo.User.UserID.String()

	res, err := h.services.User.AcceptTerms(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Changes the password for the current user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        password   body      user_service.ChangePasswordRequest                         true  "Password change request"
// @Success      204      {object}  nil
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /user/password [patch]
func (h *Handler) ChangePassword(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	var req user_service.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}

	req.UserID = accessInfo.User.UserID.String()

	if err := h.services.User.ChangePassword(&req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// GetCurrentUser godoc
// @Summary      Get current user
// @Description  Get current user details
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200      {object}  user_service.GetResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /user [get]
func (h *Handler) GetCurrentUser(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}

	resp, err := h.services.User.Get(&user_service.GetRequest{
		UserID: accessInfo.User.UserID.String(),
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// ResendVerificationEmail godoc
// @Summary      Resend verification email
// @Description  Resends email verification email to the current user
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      409      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      429      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /user/resend-verification-email [post]
func (h *Handler) ResendVerificationEmail(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}

	if err := h.services.Auth.ResendVerificationEmail(&auth_service.ResendVerificationEmailRequest{
		UserID: accessInfo.User.UserID.String(),
	}); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateProfileLanguage godoc
// @Summary      Update profile language
// @Description  Updates the language for the current user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        language   body      user_service.UpdateLanguageRequest  true  "Language update request"
// @Success      200      {object}  models.User
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /user/language [patch]
func (h *Handler) UpdateProfileLanguage(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}

	var req user_service.UpdateLanguageRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	req.UserID = accessInfo.User.UserID.String()

	res, err := h.services.User.UpdateLanguage(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// UpdateProfile godoc
// @Summary      Update profile
// @Description  Updates the profile for the current user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        profile   body      user_service.UpdateProfileRequest  true  "Profile update request"
// @Success      200      {object}  models.User
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /user/profile [patch]
func (h *Handler) UpdateProfile(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewUnauthorizedError(nil, "", nil)
	}

	var req user_service.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	req.UserID = accessInfo.User.UserID.String()

	res, err := h.services.User.UpdateProfile(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
