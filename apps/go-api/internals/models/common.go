package models

import (
	"time"

	"github.com/google/uuid"
)

type PaginatedListResponse[T any] struct {
	Items   []T  `json:"items"`
	HasNext bool `json:"hasNext"`
}

type JWTAccessTokenPermissionData struct {
	OrganizationID uuid.UUID       `json:"organizationId"`
	RoleID         uuid.UUID       `json:"roleId"`
	Permissions    RolePermissions `json:"permissions"`
}

type JWTAccessTokenClaims struct {
	Subject             string                         `json:"sub"`
	Email               string                         `json:"email"`
	ExpiresAt           time.Time                      `json:"exp"`
	IssuedAt            time.Time                      `json:"iat"`
	ImpersonatedBy      uuid.NullUUID                  `json:"impersonatedBy,omitempty"`
	ImpersonatedByEmail string                         `json:"impersonatedByEmail,omitempty"`
	Permissions         []JWTAccessTokenPermissionData `json:"permissions"`
	IsAdmin             bool                           `json:"isAdmin,omitempty"`
}

type UserAccessInfo struct {
	UserID              uuid.UUID
	Email               string
	ImpersonatedBy      uuid.NullUUID
	ImpersonatedByEmail string
	Roles               []JWTAccessTokenPermissionData
	IsAdmin             bool
}

type AccessInfo struct {
	RequestID string
	IPAddress string
	UserAgent string
	User      *UserAccessInfo
	ApiKey    *APIKey
}

// ErrorResponse is the API error body shape used for Swagger/OpenAPI documentation.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
