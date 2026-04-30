package auth_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	auth_service "github.com/yca-software/2chi-kit/go-api/internals/services/auth"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_http "github.com/yca-software/go-common/http"
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
	group := router.Group("/auth")

	authRateLimit := http_middlewares.AuthRateLimitMiddleware(h.rateLimitConfig)
	passwordResetRateLimit := http_middlewares.PasswordResetRateLimitMiddleware(h.rateLimitConfig)

	group.POST("/oauth/google", h.AuthenticateWithGoogle, authRateLimit)
	group.POST("/login", h.AuthenticateWithPassword, authRateLimit)
	group.POST("/logout", h.Logout, authRateLimit, h.middlewares.PrivateRoute)
	group.POST("/forgot-password", h.ForgotPassword, passwordResetRateLimit)
	group.POST("/refresh", h.RefreshAccessToken, authRateLimit)
	group.POST("/reset-password", h.ResetPassword, passwordResetRateLimit)
	group.POST("/signup", h.SignUp, authRateLimit)
	group.POST("/verify-email", h.VerifyEmail, authRateLimit)
}

// AuthenticateWithGoogle godoc
// @Summary      Authenticate user with Google OAuth
// @Description  Login or register with Google OAuth, returning access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth_service.AuthenticateWithGoogleRequest  true  "Google OAuth request"
// @Success      200      {object}  auth_service.AuthenticateResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      429      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/oauth/google [post]
func (h *Handler) AuthenticateWithGoogle(c echo.Context) error {
	var req auth_service.AuthenticateWithGoogleRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	req.UserAgent = c.Request().UserAgent()
	req.IPAddress = c.RealIP()
	req.Language = yca_http.GetLanguage(c, constants.SUPPORTED_LANGUAGES, constants.DEFAULT_LANGUAGE)

	resp, err := h.services.Auth.AuthenticateWithGoogle(&req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// AuthenticateWithPassword godoc
// @Summary      Authenticate user with email and password
// @Description  Login with email and password to receive access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth_service.AuthenticateWithPasswordRequest  true  "Login request"
// @Success      200      {object}  auth_service.AuthenticateResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      429      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/login [post]
func (h *Handler) AuthenticateWithPassword(c echo.Context) error {
	var req auth_service.AuthenticateWithPasswordRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	req.UserAgent = c.Request().UserAgent()
	req.IPAddress = c.RealIP()

	resp, err := h.services.Auth.AuthenticateWithPassword(&req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// ForgotPassword godoc
// @Summary      Request password reset
// @Description  Send a password reset email to the user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth_service.ForgotPasswordRequest  true  "Forgot password request"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      429      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c echo.Context) error {
	var req auth_service.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	req.Language = yca_http.GetLanguage(c, constants.SUPPORTED_LANGUAGES, constants.DEFAULT_LANGUAGE)

	if err := h.services.Auth.ForgotPassword(&req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// Logout godoc
// @Summary      Logout user
// @Description  Invalidate the refresh token to log out the user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Param        request  body      auth_service.LogoutRequest  true  "Logout request"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/logout [post]
func (h *Handler) Logout(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	var req auth_service.LogoutRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.services.Auth.Logout(&req, accessInfo); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// RefreshAccessToken godoc
// @Summary      Refresh access token
// @Description  Get a new access token using a valid refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth_service.RefreshAccessTokenRequest  true  "Refresh token request"
// @Success      200      {object}  auth_service.RefreshAccessTokenResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      429      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/refresh [post]
func (h *Handler) RefreshAccessToken(c echo.Context) error {
	var req auth_service.RefreshAccessTokenRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	req.UserAgent = c.Request().UserAgent()
	req.IPAddress = c.RealIP()

	resp, err := h.services.Auth.RefreshAccessToken(&req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Reset user password using the reset token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth_service.ResetPasswordRequest  true  "Reset password request"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      429      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/reset-password [post]
func (h *Handler) ResetPassword(c echo.Context) error {
	var req auth_service.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.services.Auth.ResetPassword(&req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// SignUp godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth_service.SignUpRequest  true  "Sign up request"
// @Success      201      {object}  auth_service.SignUpResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      409      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      429      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/signup [post]
func (h *Handler) SignUp(c echo.Context) error {
	var req auth_service.SignUpRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	req.Language = yca_http.GetLanguage(c, constants.SUPPORTED_LANGUAGES, constants.DEFAULT_LANGUAGE)
	req.UserAgent = c.Request().UserAgent()
	req.IPAddress = c.RealIP()

	resp, err := h.services.Auth.SignUp(&req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, resp)
}

// VerifyEmail godoc
// @Summary      Verify email address
// @Description  Verify user email address using verification token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth_service.VerifyEmailRequest  true  "Verify email request"
// @Success      204      "No Content"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      429      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/verify-email [post]
func (h *Handler) VerifyEmail(c echo.Context) error {
	var req auth_service.VerifyEmailRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.services.Auth.VerifyEmail(&req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
