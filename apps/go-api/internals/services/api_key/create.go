package api_key_service

import (
	"encoding/json"
	"time"

	"slices"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type CreateRequest struct {
	OrganizationID string                 `json:"organizationId" validate:"required,uuid"`
	Name           string                 `json:"name" validate:"required,min=1,max=255"`
	Permissions    models.RolePermissions `json:"permissions" validate:"required,min=1"`
	ExpiresAt      time.Time              `json:"expiresAt" validate:"required"`
}

type CreateResponse struct {
	ApiKey *models.APIKey `json:"apiKey"`
	Secret string         `json:"secret"` // The raw key, shown only once at creation
}

func (s *service) Create(req *CreateRequest, accessInfo *models.AccessInfo) (*CreateResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	for _, permission := range req.Permissions {
		if !slices.Contains(assignablePermissions, permission) {
			return nil, yca_error.NewUnprocessableEntityError(nil, constants.INVALID_API_KEY_PERMISSION_CODE, nil)
		}
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_API_KEY_WRITE); err != nil {
		return nil, err
	}
	if err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_API_ACCESS); err != nil {
		return nil, err
	}

	rawKey, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	keyPrefix := constants.API_KEY_PREFIX + rawKey[:constants.API_KEY_PREFIX_LEN]

	apiKeyID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	now := s.now()
	apiKey := &models.APIKey{
		ID:             apiKeyID,
		CreatedAt:      now,
		OrganizationID: org.ID,
		Name:           req.Name,
		KeyPrefix:      keyPrefix,
		KeyHash:        s.hashToken(rawKey),
		Permissions:    req.Permissions,
		ExpiresAt:      req.ExpiresAt,
	}

	if err := s.repos.ApiKey.Create(nil, apiKey); err != nil {
		return nil, err
	}

	auditPayload, err := json.Marshal(map[string]any{
		"name":        req.Name,
		"permissions": req.Permissions,
		"expiresAt":   req.ExpiresAt,
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.api_keyService.Create",
			Error:    err,
			Message:  "Failed to marshal audit payload",
			Data:     map[string]any{"organization_id": req.OrganizationID},
		})
	} else {
		auditRaw := json.RawMessage(auditPayload)
		nameCopy := apiKey.Name
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: req.OrganizationID,
			Action:         constants.AUDIT_ACTION_TYPE_CREATE,
			ResourceType:   constants.RESOURCE_TYPE_API_KEY,
			ResourceID:     apiKey.ID.String(),
			ResourceName:   &nameCopy,
			Data:           &auditRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.api_keyService.Create",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID},
			})
		}
	}

	return &CreateResponse{
		ApiKey: apiKey,
		Secret: constants.API_KEY_PREFIX + rawKey,
	}, nil
}
