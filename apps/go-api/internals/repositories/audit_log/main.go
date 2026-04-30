package audit_log_repository

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_observe "github.com/yca-software/go-common/observer"
	yca_repository "github.com/yca-software/go-common/repository"
)

var (
	TABLE_NAME = "audit_logs"
	COLUMNS    = []string{"id", "created_at", "organization_id", "actor_id", "actor_info", "impersonated_by_id", "impersonated_by_email", "action", "resource_type", "resource_id", "resource_name", "data"}
)

type AuditLogFilters struct {
	StartDate *time.Time
	EndDate   *time.Time
}

type Repository interface {
	yca_repository.Repository[models.AuditLog]
	Create(tx yca_repository.Tx, log *models.AuditLog) error
	ListByOrganizationID(orgID string, filters *AuditLogFilters, limit, offset int) (*[]models.AuditLog, error)
}

type repository struct {
	yca_repository.Repository[models.AuditLog]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.AuditLog](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, log *models.AuditLog) error {
	return r.BaseCreate(tx, map[string]any{
		"id":                    log.ID,
		"created_at":            log.CreatedAt,
		"organization_id":       log.OrganizationID,
		"actor_id":              log.ActorID,
		"actor_info":            log.ActorInfo,
		"impersonated_by_id":    log.ImpersonatedByID,
		"impersonated_by_email": log.ImpersonatedByEmail,
		"action":                log.Action,
		"resource_type":         log.ResourceType,
		"resource_id":           log.ResourceID,
		"resource_name":         log.ResourceName,
		"data":                  log.Data,
	})
}

func (r *repository) ListByOrganizationID(orgID string, filters *AuditLogFilters, limit, offset int) (*[]models.AuditLog, error) {
	condition := squirrel.And{squirrel.Eq{"organization_id": orgID}}
	if filters != nil {
		if filters.StartDate != nil {
			condition = append(condition, squirrel.GtOrEq{"created_at": filters.StartDate})
		}
		if filters.EndDate != nil {
			condition = append(condition, squirrel.Lt{"created_at": filters.EndDate})
		}
	}
	return r.BasePaginatedSelect(nil, condition, nil, "created_at DESC", uint64(limit), uint64(offset))
}
