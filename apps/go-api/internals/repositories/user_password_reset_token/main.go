package user_password_reset_token_repository

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
	TABLE_NAME = "user_password_reset_tokens"
	COLUMNS    = []string{"id", "user_id", "created_at", "expires_at", "used_at", "token_hash"}
)

type Repository interface {
	yca_repository.Repository[models.UserPasswordResetToken]
	Create(tx yca_repository.Tx, token *models.UserPasswordResetToken) error
	Cleanup(tx yca_repository.Tx) error
	GetByHash(tx yca_repository.Tx, tokenHash string) (*models.UserPasswordResetToken, error)
	MarkAsUsed(tx yca_repository.Tx, tokenID string) error
}

type repository struct {
	yca_repository.Repository[models.UserPasswordResetToken]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.UserPasswordResetToken](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, token *models.UserPasswordResetToken) error {
	return r.BaseCreate(tx, map[string]any{
		"id":         token.ID,
		"user_id":    token.UserID,
		"created_at": token.CreatedAt,
		"expires_at": token.ExpiresAt,
		"token_hash": token.TokenHash,
	})
}

// Cleanup deletes password reset token rows at least StaleDataRetentionPeriod after they became unused (used or past expires_at).
func (r *repository) Cleanup(tx yca_repository.Tx) error {
	threshold := time.Now().Add(-constants.StaleDataRetentionPeriod)
	return r.BaseDelete(tx, squirrel.Or{
		squirrel.And{
			squirrel.NotEq{"used_at": nil},
			squirrel.LtOrEq{"used_at": threshold},
		},
		squirrel.And{
			squirrel.Eq{"used_at": nil},
			squirrel.LtOrEq{"expires_at": threshold},
		},
	})
}

func (r *repository) GetByHash(tx yca_repository.Tx, tokenHash string) (*models.UserPasswordResetToken, error) {
	return r.BaseGet(tx, squirrel.Eq{"token_hash": tokenHash}, nil)
}

func (r *repository) MarkAsUsed(tx yca_repository.Tx, tokenID string) error {
	return r.BaseUpdate(tx, squirrel.Eq{"id": tokenID, "used_at": nil}, map[string]any{"used_at": time.Now()})
}
