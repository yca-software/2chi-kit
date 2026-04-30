package http_middlewares

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

func (m *middlewares) PrivateRoute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := c.Response().Header().Get(echo.HeaderXRequestID)
		ipAddress := c.RealIP()
		userAgent := c.Request().UserAgent()
		apiKeyHeader := c.Request().Header.Get("X-API-Key")

		if apiKeyHeader != "" {
			rawKey := strings.TrimPrefix(apiKeyHeader, constants.API_KEY_PREFIX)
			keyHash := m.hashToken(rawKey)

			apiKey, err := m.repos.ApiKey.GetByHash(keyHash)
			if err != nil {
				if e, ok := yca_error.AsError(err); ok && e.ErrorCode == constants.NOT_FOUND_CODE {
					return yca_error.NewUnauthorizedError(nil, constants.INVALID_API_KEY_CODE, nil)
				}
				return err
			}

			if apiKey.ExpiresAt.Before(time.Now()) {
				return yca_error.NewUnauthorizedError(nil, constants.EXPIRED_TOKEN_CODE, nil)
			}

			org, err := m.repos.Organization.GetByID(apiKey.OrganizationID.String())
			if err != nil {
				return err
			}
			allowedTypes := constants.FEATURES_FOR_PLANS[constants.FEATURE_API_ACCESS]
			if !slices.Contains(allowedTypes, org.SubscriptionType) {
				return yca_error.NewForbiddenError(nil, constants.FEATURE_NOT_INCLUDED_CODE, nil)
			}

			accessInfo := &models.AccessInfo{
				RequestID: requestID,
				IPAddress: ipAddress,
				UserAgent: userAgent,
				ApiKey:    apiKey,
			}
			c.Set("accessInfo", accessInfo)
			return next(c)
		} else {
			tokenString := c.Request().Header.Get("Authorization")
			if tokenString == "" {
				tokenString = "Bearer " + c.Request().Header.Get("X-API-Key")
			}
			if tokenString == "" {
				return yca_error.NewUnauthorizedError(nil, constants.AUTHORIZATION_HEADER_REQUIRED_CODE, nil)
			}
			if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
				return yca_error.NewUnauthorizedError(nil, constants.AUTHORIZATION_HEADER_REQUIRED_CODE, nil)
			}
			tokenString = strings.TrimSpace(tokenString[7:])
			if tokenString == "" {
				return yca_error.NewUnauthorizedError(nil, constants.AUTHORIZATION_HEADER_REQUIRED_CODE, nil)
			}

			tokenClaims, err := m.validateAccessToken(tokenString)
			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					return yca_error.NewUnauthorizedError(err, constants.EXPIRED_TOKEN_CODE, nil)
				}
				if errors.Is(err, jwt.ErrSignatureInvalid) || errors.Is(err, jwt.ErrInvalidKey) {
					return yca_error.NewUnauthorizedError(err, constants.INVALID_TOKEN_CODE, nil)
				}
				return yca_error.NewUnauthorizedError(err, constants.INVALID_TOKEN_CODE, nil)
			}

			userID, err := uuid.Parse(tokenClaims.Subject)
			if err != nil {
				return yca_error.NewUnauthorizedError(err, constants.INVALID_TOKEN_CODE, nil)
			}
			accessInfo := &models.AccessInfo{
				RequestID: requestID,
				IPAddress: ipAddress,
				UserAgent: userAgent,
				User: &models.UserAccessInfo{
					UserID:              userID,
					Email:               tokenClaims.Email,
					ImpersonatedBy:      tokenClaims.ImpersonatedBy,
					ImpersonatedByEmail: tokenClaims.ImpersonatedByEmail,
					Roles:               tokenClaims.Permissions,
					IsAdmin:             tokenClaims.IsAdmin,
				},
			}
			c.Set("accessInfo", accessInfo)
			return next(c)
		}
	}
}
