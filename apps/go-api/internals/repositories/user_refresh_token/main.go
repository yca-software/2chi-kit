package user_refresh_token_repository

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
	TABLE_NAME = "user_refresh_tokens"
	COLUMNS    = []string{"id", "user_id", "created_at", "expires_at", "revoked_at", "ip", "user_agent", "token_hash", "impersonated_by"}
)

type Repository interface {
	yca_repository.Repository[models.UserRefreshToken]
	Create(tx yca_repository.Tx, token *models.UserRefreshToken) error
	CleanupStaleUnused(tx yca_repository.Tx) error
	GetByHash(tx yca_repository.Tx, tokenHash string) (*models.UserRefreshToken, error)
	GetActiveByUserID(tx yca_repository.Tx, userID string) (*[]models.UserRefreshToken, error)
	GetActiveImpersonationTokenByUserID(tx yca_repository.Tx, userID string) (*models.UserRefreshToken, error)
	Revoke(tx yca_repository.Tx, userID string, tokenID string) error
	RevokeByHash(tx yca_repository.Tx, tokenHash string) error
	RevokeAll(tx yca_repository.Tx, userID string) error
	RevokeAllExcept(tx yca_repository.Tx, userID string, excludeTokenID string) error
}

type repository struct {
	yca_repository.Repository[models.UserRefreshToken]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.UserRefreshToken](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, token *models.UserRefreshToken) error {
	return r.BaseCreate(tx, map[string]any{
		"id":              token.ID,
		"user_id":         token.UserID,
		"created_at":      token.CreatedAt,
		"expires_at":      token.ExpiresAt,
		"ip":              token.IP,
		"user_agent":      token.UserAgent,
		"token_hash":      token.TokenHash,
		"impersonated_by": token.ImpersonatedBy,
	})
}

// CleanupStaleUnused deletes refresh token rows at least StaleDataRetentionPeriod after they became unused (revoked or past expires_at).
func (r *repository) CleanupStaleUnused(tx yca_repository.Tx) error {
	threshold := time.Now().Add(-constants.StaleDataRetentionPeriod)
	return r.BaseDelete(tx, squirrel.Or{
		squirrel.And{
			squirrel.NotEq{"revoked_at": nil},
			squirrel.LtOrEq{"revoked_at": threshold},
		},
		squirrel.And{
			squirrel.Eq{"revoked_at": nil},
			squirrel.LtOrEq{"expires_at": threshold},
		},
	})
}

func (r *repository) GetByHash(tx yca_repository.Tx, tokenHash string) (*models.UserRefreshToken, error) {
	return r.BaseGet(tx, squirrel.Eq{"token_hash": tokenHash}, nil)
}

func (r *repository) GetActiveByUserID(tx yca_repository.Tx, userID string) (*[]models.UserRefreshToken, error) {
	return r.BaseSelect(tx, squirrel.And{
		squirrel.Eq{"user_id": userID},
		squirrel.Eq{"revoked_at": nil},
		squirrel.Gt{"expires_at": time.Now()},
		squirrel.Eq{"impersonated_by": nil},
	}, nil, "created_at DESC")
}

func (r *repository) GetActiveImpersonationTokenByUserID(tx yca_repository.Tx, userID string) (*models.UserRefreshToken, error) {
	return r.BaseGet(tx, squirrel.And{
		squirrel.Eq{"user_id": userID},
		squirrel.Eq{"revoked_at": nil},
		squirrel.Gt{"expires_at": time.Now()},
		squirrel.NotEq{"impersonated_by": nil},
	}, nil)
}

func (r *repository) Revoke(tx yca_repository.Tx, userID string, tokenID string) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"user_id": userID},
		squirrel.Eq{"id": tokenID},
		squirrel.Eq{"revoked_at": nil},
		squirrel.Gt{"expires_at": time.Now()},
	}, map[string]any{"revoked_at": time.Now()})
}

func (r *repository) RevokeByHash(tx yca_repository.Tx, tokenHash string) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"token_hash": tokenHash},
		squirrel.Eq{"revoked_at": nil},
		squirrel.Gt{"expires_at": time.Now()},
	}, map[string]any{"revoked_at": time.Now()})
}

func (r *repository) RevokeAll(tx yca_repository.Tx, userID string) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"user_id": userID},
		squirrel.Eq{"revoked_at": nil},
		squirrel.Gt{"expires_at": time.Now()},
	}, map[string]any{"revoked_at": time.Now()})
}

func (r *repository) RevokeAllExcept(tx yca_repository.Tx, userID string, excludeTokenID string) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"user_id": userID},
		squirrel.NotEq{"id": excludeTokenID},
		squirrel.Eq{"revoked_at": nil},
		squirrel.Gt{"expires_at": time.Now()},
	}, map[string]any{"revoked_at": time.Now()})
}
