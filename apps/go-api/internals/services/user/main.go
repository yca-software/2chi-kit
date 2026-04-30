package user_service

import (
	"time"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	GenerateID     func() (uuid.UUID, error)
	Now            func() time.Time
	Validator      yca_validate.Validator
	Repositories   *repositories.Repositories
	Authorizer     *helpers.Authorizer
	Logger         yca_log.Logger
	PasswordHashFn func(password string) (string, error)
	GenerateToken  func() (string, error)
	HashToken      func(token string) string
}

type Service interface {
	AcceptTerms(req *AcceptTermsRequest, accessInfo *models.AccessInfo) (*models.User, error)
	ChangePassword(req *ChangePasswordRequest) error
	Count(accessInfo *models.AccessInfo) (int, error)
	Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error
	Get(req *GetRequest, accessInfo *models.AccessInfo) (*GetResponse, error)
	List(req *ListRequest, accessInfo *models.AccessInfo) (*PaginatedListResponse, error)
	UpdateProfile(req *UpdateProfileRequest, accessInfo *models.AccessInfo) (*models.User, error)
	UpdateLanguage(req *UpdateLanguageRequest, accessInfo *models.AccessInfo) (*models.User, error)
}

type service struct {
	generateID     func() (uuid.UUID, error)
	now            func() time.Time
	validator      yca_validate.Validator
	repos          *repositories.Repositories
	authorizer     *helpers.Authorizer
	logger         yca_log.Logger
	passwordHashFn func(password string) (string, error)
	generateToken  func() (string, error)
	hashToken      func(token string) string
}

func NewService(deps *Dependencies) Service {
	return &service{
		generateID:     deps.GenerateID,
		now:            deps.Now,
		validator:      deps.Validator,
		repos:          deps.Repositories,
		authorizer:     deps.Authorizer,
		logger:         deps.Logger,
		passwordHashFn: deps.PasswordHashFn,
		generateToken:  deps.GenerateToken,
		hashToken:      deps.HashToken,
	}
}
