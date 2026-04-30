package user_refresh_token_service

import (
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	Now          func() time.Time
	Validator    yca_validate.Validator
	Repositories *repositories.Repositories
	Authorizer   *helpers.Authorizer
	Logger       yca_log.Logger
	HashToken    func(token string) string
}

type Service interface {
	Revoke(req *RevokeRequest, accessInfo *models.AccessInfo) error
	RevokeAll(req *RevokeAllRequest, accessInfo *models.AccessInfo) error
	ListActive(req *ListActiveRequest, accessInfo *models.AccessInfo) (*[]models.UserRefreshToken, error)
	CleanupStaleUnused() error
}

type service struct {
	now        func() time.Time
	validator  yca_validate.Validator
	repos      *repositories.Repositories
	authorizer *helpers.Authorizer
	logger     yca_log.Logger
	hashToken  func(token string) string
}

func NewService(deps *Dependencies) Service {
	return &service{
		now:        deps.Now,
		validator:  deps.Validator,
		repos:      deps.Repositories,
		authorizer: deps.Authorizer,
		logger:     deps.Logger,
		hashToken:  deps.HashToken,
	}
}
