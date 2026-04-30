package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`

	OrganizationID uuid.UUID `db:"organization_id" json:"organizationId"`

	ActorID             uuid.UUID     `db:"actor_id" json:"actorId"`
	ActorInfo           string        `db:"actor_info" json:"actorInfo"`
	ImpersonatedByID    uuid.NullUUID `db:"impersonated_by_id" json:"impersonatedById"`
	ImpersonatedByEmail string        `db:"impersonated_by_email" json:"impersonatedByEmail"`

	Action       string    `db:"action" json:"action"`
	ResourceType string    `db:"resource_type" json:"resourceType"`
	ResourceID   uuid.UUID `db:"resource_id" json:"resourceId"`
	ResourceName *string   `db:"resource_name" json:"resourceName,omitempty"`

	Data *json.RawMessage `db:"data" json:"data"`
}

// AuditLogPublic is the audit log shape returned to API clients (sanitized).
type AuditLogPublic struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	OrganizationID uuid.UUID `json:"organizationId"`

	ActorID             uuid.UUID  `json:"actorId"`
	ActorInfo           string     `json:"actorInfo"`
	ImpersonatedByID    *uuid.UUID `json:"impersonatedById,omitempty"`
	ImpersonatedByEmail string     `json:"impersonatedByEmail,omitempty"`

	Action       string    `json:"action"`
	ResourceType string    `json:"resourceType"`
	ResourceID   uuid.UUID `json:"resourceId"`
	ResourceName *string   `json:"resourceName,omitempty"`

	Data json.RawMessage `json:"data"`
}
