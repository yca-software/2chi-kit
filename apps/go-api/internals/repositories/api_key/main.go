package api_key_repository

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_observe "github.com/yca-software/go-common/observer"
	yca_repository "github.com/yca-software/go-common/repository"
)

var (
	TABLE_NAME = "api_keys"
	COLUMNS    = []string{"id", "created_at", "name", "key_prefix", "key_hash", "organization_id", "permissions", "expires_at"}
)

type Repository interface {
	yca_repository.Repository[models.APIKey]
	Create(tx yca_repository.Tx, apiKey *models.APIKey) error
	GetByHash(keyHash string) (*models.APIKey, error)
	GetByID(organizationID string, id string) (*models.APIKey, error)
	ListByOrganizationID(organizationID string) (*[]models.APIKey, error)
	Update(tx yca_repository.Tx, apiKey *models.APIKey) error
	Delete(tx yca_repository.Tx, organizationID string, id string) error
	CleanupStaleExpired() error
}

type repository struct {
	yca_repository.Repository[models.APIKey]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.APIKey](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, apiKey *models.APIKey) error {
	return r.BaseCreate(tx, map[string]any{
		"id":              apiKey.ID,
		"created_at":      apiKey.CreatedAt,
		"name":            apiKey.Name,
		"key_prefix":      apiKey.KeyPrefix,
		"key_hash":        apiKey.KeyHash,
		"organization_id": apiKey.OrganizationID,
		"permissions":     apiKey.Permissions,
		"expires_at":      apiKey.ExpiresAt,
	})
}

func (r *repository) GetByHash(keyHash string) (*models.APIKey, error) {
	return r.BaseGet(nil, squirrel.And{
		squirrel.Eq{"key_hash": keyHash},
	}, nil)
}

func (r *repository) GetByID(organizationID string, id string) (*models.APIKey, error) {
	return r.BaseGet(nil, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	}, nil)
}

func (r *repository) ListByOrganizationID(organizationID string) (*[]models.APIKey, error) {
	return r.BaseSelect(nil, squirrel.And{
		squirrel.Eq{"organization_id": organizationID},
	}, nil, "created_at DESC")
}

func (r *repository) Update(tx yca_repository.Tx, apiKey *models.APIKey) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"id": apiKey.ID},
		squirrel.Eq{"organization_id": apiKey.OrganizationID},
	}, map[string]any{
		"name":        apiKey.Name,
		"permissions": apiKey.Permissions,
	})
}

func (r *repository) Delete(tx yca_repository.Tx, organizationID string, id string) error {
	return r.BaseDelete(tx, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	})
}

// CleanupStaleExpired deletes API key rows with a set expires_at at least StaleDataRetentionPeriod after expiry. Keys without expires_at are kept.
func (r *repository) CleanupStaleExpired() error {
	threshold := time.Now().Add(-constants.StaleDataRetentionPeriod)
	return r.BaseDelete(nil, squirrel.And{
		squirrel.Expr("expires_at IS NOT NULL"),
		squirrel.LtOrEq{"expires_at": threshold},
	})
}
