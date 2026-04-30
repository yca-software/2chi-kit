package team_repository

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_observe "github.com/yca-software/go-common/observer"
	yca_repository "github.com/yca-software/go-common/repository"
)

var (
	TABLE_NAME = "teams"
	COLUMNS    = []string{"id", "created_at", "organization_id", "name", "description"}
)

type Repository interface {
	yca_repository.Repository[models.Team]
	Create(tx yca_repository.Tx, team *models.Team) error
	GetByID(organizationID string, id string) (*models.Team, error)
	ListByOrganizationID(organizationID string) (*[]models.Team, error)
	Update(tx yca_repository.Tx, team *models.Team) error
	Delete(tx yca_repository.Tx, organizationID string, id string) error
}

type repository struct {
	yca_repository.Repository[models.Team]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.Team](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, team *models.Team) error {
	return r.BaseCreate(tx, map[string]any{
		"id":              team.ID,
		"created_at":      team.CreatedAt,
		"organization_id": team.OrganizationID,
		"name":            team.Name,
		"description":     team.Description,
	})
}

func (r *repository) GetByID(organizationID string, id string) (*models.Team, error) {
	return r.BaseGet(nil, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	}, nil)
}

func (r *repository) Update(tx yca_repository.Tx, team *models.Team) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"id": team.ID},
		squirrel.Eq{"organization_id": team.OrganizationID},
	}, map[string]any{
		"name":        team.Name,
		"description": team.Description,
	})
}

func (r *repository) Delete(tx yca_repository.Tx, organizationID string, id string) error {
	return r.BaseDelete(tx, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	})
}

func (r *repository) ListByOrganizationID(organizationID string) (*[]models.Team, error) {
	return r.BaseSelect(nil, squirrel.Eq{"organization_id": organizationID}, nil, "name ASC")
}
