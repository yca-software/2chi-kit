package audit_log_service

import (
	"time"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	Validator  yca_validate.Validator
	Repos      *repositories.Repositories
	Logger     yca_log.Logger
	GenerateID func() (uuid.UUID, error)
	Now        func() time.Time
	Authorizer *helpers.Authorizer
}

type Service interface {
	Create(req *CreateRequest, accessInfo *models.AccessInfo) (*models.AuditLog, error)
	ListForOrganization(req *ListForOrganizationRequest, accessInfo *models.AccessInfo) (*ListForOrganizationResponse, error)
}

type service struct {
	validator  yca_validate.Validator
	repos      *repositories.Repositories
	logger     yca_log.Logger
	generateID func() (uuid.UUID, error)
	now        func() time.Time
	authorizer *helpers.Authorizer
}

func New(deps *Dependencies) Service {
	return &service{
		validator:  deps.Validator,
		repos:      deps.Repos,
		logger:     deps.Logger,
		generateID: deps.GenerateID,
		now:        deps.Now,
		authorizer: deps.Authorizer,
	}
}
