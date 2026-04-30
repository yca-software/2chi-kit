package role_repository

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_observe "github.com/yca-software/go-common/observer"
	yca_repository "github.com/yca-software/go-common/repository"
)

var (
	TABLE_NAME = "roles"
	COLUMNS    = []string{"id", "created_at", "organization_id", "name", "description", "permissions", "locked"}
)

type Repository interface {
	yca_repository.Repository[models.Role]
	Create(tx yca_repository.Tx, role *models.Role) error
	CreateMany(tx yca_repository.Tx, roles *[]models.Role) error
	Delete(tx yca_repository.Tx, organizationID string, id string) error
	GetByID(organizationID string, id string) (*models.Role, error)
	ListByOrganizationID(organizationID string) (*[]models.Role, error)
	Update(tx yca_repository.Tx, role *models.Role) error
}

type repository struct {
	yca_repository.Repository[models.Role]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.Role](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, role *models.Role) error {
	return r.BaseCreate(tx, map[string]any{
		"id":              role.ID,
		"created_at":      role.CreatedAt,
		"organization_id": role.OrganizationID,
		"name":            role.Name,
		"description":     role.Description,
		"permissions":     role.Permissions,
		"locked":          role.Locked,
	})
}

func (r *repository) GetByID(organizationID string, id string) (*models.Role, error) {
	return r.BaseGet(nil, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	}, nil)
}

func (r *repository) Update(tx yca_repository.Tx, role *models.Role) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"id": role.ID},
		squirrel.Eq{"organization_id": role.OrganizationID},
		squirrel.Eq{"locked": false},
	}, map[string]any{
		"name":        role.Name,
		"description": role.Description,
		"permissions": role.Permissions,
	})
}

func (r *repository) ListByOrganizationID(organizationID string) (*[]models.Role, error) {
	return r.BaseSelect(nil, squirrel.Eq{"organization_id": organizationID}, nil, "name ASC")
}

func (r *repository) Delete(tx yca_repository.Tx, organizationID string, id string) error {
	return r.BaseDelete(tx, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
		squirrel.Eq{"locked": false},
	})
}

func (r *repository) CreateMany(tx yca_repository.Tx, roles *[]models.Role) error {
	data := make([]map[string]any, len(*roles))
	for i, role := range *roles {
		data[i] = map[string]any{
			"id":              role.ID,
			"created_at":      role.CreatedAt,
			"organization_id": role.OrganizationID,
			"name":            role.Name,
			"description":     role.Description,
			"permissions":     role.Permissions,
			"locked":          role.Locked,
		}
	}
	return r.BaseCreateMany(tx, COLUMNS, data, false)
}
