package audit_log_service

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type CreateRequest struct {
	OrganizationID string           `json:"organizationId" validate:"required,uuid"`
	Action         string           `json:"action" validate:"required"`
	ResourceType   string           `json:"resourceType" validate:"required"`
	ResourceID     string           `json:"resourceId" validate:"required,uuid"`
	ResourceName   *string          `json:"resourceName" validate:"omitempty,max=512"`
	Data           *json.RawMessage `json:"data" validate:"required"` // optional for archive/restore via custom handling if needed
}

func (s *service) Create(req *CreateRequest, accessInfo *models.AccessInfo) (*models.AuditLog, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	auditLogID, err := s.generateID()
	if err != nil {
		return nil, err
	}
	now := s.now()
	var actorID uuid.UUID
	var actorInfo string
	if accessInfo.User != nil {
		actorID = accessInfo.User.UserID
		actorInfo = accessInfo.User.Email
	} else if accessInfo.ApiKey != nil {
		actorID = accessInfo.ApiKey.ID
		actorInfo = accessInfo.ApiKey.KeyPrefix
	}

	var data *json.RawMessage
	if req.Data != nil {
		sanitized := SanitizeAuditDataJSON(*req.Data)
		data = &sanitized
	}

	auditLog := models.AuditLog{
		ID:             auditLogID,
		CreatedAt:      now,
		OrganizationID: uuid.MustParse(req.OrganizationID),
		ActorID:        actorID,
		ActorInfo:      actorInfo,
		Action:         req.Action,
		ResourceType:   req.ResourceType,
		ResourceID:     uuid.MustParse(req.ResourceID),
		ResourceName:   req.ResourceName,
		Data:           data,
	}

	if accessInfo.User != nil {
		auditLog.ImpersonatedByID = accessInfo.User.ImpersonatedBy
		if accessInfo.User.ImpersonatedBy.Valid {
			auditLog.ImpersonatedByEmail = accessInfo.User.ImpersonatedByEmail
		}
	}

	if err := s.repos.AuditLog.Create(nil, &auditLog); err != nil {
		return nil, err
	}

	return &auditLog, nil
}
