package admin_access_repository

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_observe "github.com/yca-software/go-common/observer"
	yca_repository "github.com/yca-software/go-common/repository"
)

var (
	TABLE_NAME = "admin_access"
	COLUMNS    = []string{"user_id", "created_at"}
)

type Repository interface {
	yca_repository.Repository[models.AdminAccess]
	GetByUserID(userID string) (*models.AdminAccess, error)
}

type repository struct {
	yca_repository.Repository[models.AdminAccess]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.AdminAccess](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) GetByUserID(userID string) (*models.AdminAccess, error) {
	return r.BaseGet(nil, squirrel.Eq{"user_id": userID}, nil)
}
