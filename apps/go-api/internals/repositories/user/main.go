package user_repository

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_observe "github.com/yca-software/go-common/observer"
	yca_repository "github.com/yca-software/go-common/repository"
)

var (
	TABLE_NAME = "users"
	COLUMNS    = []string{"id", "created_at", "first_name", "last_name", "language", "email", "email_verified_at", "password", "google_id", "avatar_url", "terms_accepted_at", "terms_version"}
)

type Repository interface {
	yca_repository.Repository[models.User]
	Create(tx yca_repository.Tx, user *models.User) error
	GetByEmail(tx yca_repository.Tx, email string) (*models.User, error)
	GetByID(tx yca_repository.Tx, id string) (*models.User, error)
	GetByGoogleID(tx yca_repository.Tx, googleID string) (*models.User, error)
	Update(tx yca_repository.Tx, user *models.User) error
	Delete(tx yca_repository.Tx, user *models.User) error
	Search(searchPhrase string, limit, offset int) (*[]models.User, error)
	Count() (int, error)
}

type repository struct {
	yca_repository.Repository[models.User]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.User](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, user *models.User) error {
	return r.BaseCreate(tx, map[string]any{
		"id":                user.ID,
		"created_at":        user.CreatedAt,
		"first_name":        user.FirstName,
		"last_name":         user.LastName,
		"language":          user.Language,
		"email":             user.Email,
		"email_verified_at": user.EmailVerifiedAt,
		"password":          user.Password,
		"google_id":         user.GoogleID,
		"avatar_url":        user.AvatarURL,
		"terms_accepted_at": user.TermsAcceptedAt,
		"terms_version":     user.TermsVersion,
	})
}

func (r *repository) GetByEmail(tx yca_repository.Tx, email string) (*models.User, error) {
	return r.BaseGet(tx, squirrel.Eq{"email": email}, nil)
}

func (r *repository) GetByID(tx yca_repository.Tx, id string) (*models.User, error) {
	return r.BaseGet(tx, squirrel.Eq{"id": id}, nil)
}

func (r *repository) GetByGoogleID(tx yca_repository.Tx, googleID string) (*models.User, error) {
	return r.BaseGet(tx, squirrel.Eq{"google_id": googleID}, nil)
}

func (r *repository) Update(tx yca_repository.Tx, user *models.User) error {
	return r.BaseUpdate(tx, squirrel.Eq{"id": user.ID}, map[string]any{
		"first_name":        user.FirstName,
		"last_name":         user.LastName,
		"language":          user.Language,
		"avatar_url":        user.AvatarURL,
		"email_verified_at": user.EmailVerifiedAt,
		"password":          user.Password,
		"google_id":         user.GoogleID,
		"terms_accepted_at": user.TermsAcceptedAt,
		"terms_version":     user.TermsVersion,
	})
}

func (r *repository) Delete(tx yca_repository.Tx, user *models.User) error {
	return r.BaseDelete(tx, squirrel.Eq{"id": user.ID})
}

func (r *repository) Search(searchPhrase string, limit, offset int) (*[]models.User, error) {
	condition := squirrel.And{}
	if searchPhrase != "" {
		condition = append(condition, squirrel.Or{
			squirrel.ILike{"email": "%" + searchPhrase + "%"},
			squirrel.ILike{"first_name": "%" + searchPhrase + "%"},
			squirrel.ILike{"last_name": "%" + searchPhrase + "%"},
		})
	}

	return r.BasePaginatedSelect(nil, condition, nil, "created_at DESC", uint64(limit), uint64(offset))
}

func (r *repository) Count() (int, error) {
	return r.BaseCount(nil, nil)
}
