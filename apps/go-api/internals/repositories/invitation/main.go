package invitation_repository

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
	TABLE_NAME = "invitations"
	COLUMNS    = []string{"id", "created_at", "expires_at", "accepted_at", "revoked_at", "organization_id", "role_id", "email", "invited_by_id", "invited_by_email", "token_hash"}
)

type Repository interface {
	yca_repository.Repository[models.Invitation]
	Create(tx yca_repository.Tx, invitation *models.Invitation) error
	GetByID(organizationID string, id string) (*models.Invitation, error)
	GetByTokenHash(tokenHash string) (*models.Invitation, error)
	Update(tx yca_repository.Tx, invitation *models.Invitation) error
	ListByOrganizationID(organizationID string) (*[]models.Invitation, error)
	CleanupStale() error
}

type repository struct {
	yca_repository.Repository[models.Invitation]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.Invitation](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, invitation *models.Invitation) error {
	return r.BaseCreate(tx, map[string]any{
		"id":               invitation.ID,
		"created_at":       invitation.CreatedAt,
		"expires_at":       invitation.ExpiresAt,
		"organization_id":  invitation.OrganizationID,
		"role_id":          invitation.RoleID,
		"email":            invitation.Email,
		"invited_by_id":    invitation.InvitedByID,
		"invited_by_email": invitation.InvitedByEmail,
		"token_hash":       invitation.TokenHash,
	})
}

func (r *repository) GetByID(organizationID string, id string) (*models.Invitation, error) {
	return r.BaseGet(nil, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	}, nil)
}

func (r *repository) GetByTokenHash(tokenHash string) (*models.Invitation, error) {
	return r.BaseGet(nil, squirrel.Eq{"token_hash": tokenHash}, nil)
}

func (r *repository) Update(tx yca_repository.Tx, invitation *models.Invitation) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"id": invitation.ID},
		squirrel.Eq{"organization_id": invitation.OrganizationID},
		squirrel.Eq{"accepted_at": nil},
		squirrel.Eq{"revoked_at": nil},
	}, map[string]any{
		"accepted_at": invitation.AcceptedAt,
		"revoked_at":  invitation.RevokedAt,
	})
}

func (r *repository) ListByOrganizationID(organizationID string) (*[]models.Invitation, error) {
	return r.BaseSelect(nil, squirrel.And{
		squirrel.Eq{"organization_id": organizationID},
		squirrel.Eq{"accepted_at": nil},
		squirrel.Eq{"revoked_at": nil},
	}, nil, "created_at DESC")
}

// CleanupStale deletes invitation rows at least StaleDataRetentionPeriod after they stopped being pending (accepted, revoked, or expired while still pending).
func (r *repository) CleanupStale() error {
	threshold := time.Now().Add(-constants.StaleDataRetentionPeriod)
	return r.BaseDelete(nil, squirrel.Or{
		squirrel.And{
			squirrel.NotEq{"accepted_at": nil},
			squirrel.LtOrEq{"accepted_at": threshold},
		},
		squirrel.And{
			squirrel.NotEq{"revoked_at": nil},
			squirrel.LtOrEq{"revoked_at": threshold},
		},
		squirrel.And{
			squirrel.Eq{"accepted_at": nil},
			squirrel.Eq{"revoked_at": nil},
			squirrel.LtOrEq{"expires_at": threshold},
		},
	})
}
