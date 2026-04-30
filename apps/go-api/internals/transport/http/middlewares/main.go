package http_middlewares

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/metrics"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
)

type MiddlewaresConfig struct {
	AccessSecret string
}

type Middlewares interface {
	PrivateRoute(next echo.HandlerFunc) echo.HandlerFunc
	AdminRoute(next echo.HandlerFunc) echo.HandlerFunc
}

type middlewares struct {
	config    *MiddlewaresConfig
	repos     *repositories.Repositories
	metrics   *metrics.Metrics
	hashToken func(token string) string
}

func NewMiddlewares(config *MiddlewaresConfig, repos *repositories.Repositories, metrics *metrics.Metrics, hashToken func(token string) string) Middlewares {
	return &middlewares{
		config:    config,
		repos:     repos,
		metrics:   metrics,
		hashToken: hashToken,
	}
}

func (s *middlewares) validateAccessToken(token string) (*models.JWTAccessTokenClaims, error) {
	tokenObj, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(s.config.AccessSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if !tokenObj.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	claims, ok := tokenObj.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	expF, ok := claims["exp"].(float64)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}
	exp := int64(expF)

	if exp <= time.Now().Unix() {
		return nil, jwt.ErrTokenExpired
	}

	iatF, ok := claims["iat"].(float64)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}
	iat := int64(iatF)

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	isAdmin, ok := claims["isAdmin"].(bool)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	impersonatedByID := uuid.NullUUID{Valid: false}
	impersonatedBy, ok := claims["impersonatedBy"].(string)
	if ok && impersonatedBy != "" {
		if parsed, err := uuid.Parse(impersonatedBy); err == nil {
			impersonatedByID = uuid.NullUUID{UUID: parsed, Valid: true}
		}
	}

	email, _ := claims["email"].(string)
	impersonatedByEmail, _ := claims["impersonatedByEmail"].(string)

	// JWT decodes claims from JSON, so "permissions" is []interface{}, not []JWTAccessTokenPermissionData.
	// Re-encode and decode into the target type.
	var permissions []models.JWTAccessTokenPermissionData
	if permRaw := claims["permissions"]; permRaw != nil {
		permBytes, err := json.Marshal(permRaw)
		if err != nil {
			return nil, jwt.ErrInvalidKey
		}
		if err := json.Unmarshal(permBytes, &permissions); err != nil {
			return nil, jwt.ErrInvalidKey
		}
	}

	return &models.JWTAccessTokenClaims{
		Subject:             userID,
		Email:               email,
		ExpiresAt:           time.Unix(exp, 0),
		IssuedAt:            time.Unix(iat, 0),
		ImpersonatedBy:      impersonatedByID,
		ImpersonatedByEmail: impersonatedByEmail,
		Permissions:         permissions,
		IsAdmin:             isAdmin,
	}, nil
}
