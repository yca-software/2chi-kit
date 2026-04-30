package api_key_service

import (
	"time"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	Validator       yca_validate.Validator
	Repos           *repositories.Repositories
	Authorizer      *helpers.Authorizer
	GenerateID      func() (uuid.UUID, error)
	GenerateToken   func() (string, error)
	HashToken       func(token string) string
	Now             func() time.Time
	Logger          yca_log.Logger
	AuditLogService audit_log_service.Service
}

type Service interface {
	Create(req *CreateRequest, accessInfo *models.AccessInfo) (*CreateResponse, error)
	List(req *ListRequest, accessInfo *models.AccessInfo) (*[]models.APIKey, error)
	Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error
	Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.APIKey, error)
	CleanupStaleExpired() error
}

type service struct {
	validator       yca_validate.Validator
	repos           *repositories.Repositories
	authorizer      *helpers.Authorizer
	generateID      func() (uuid.UUID, error)
	generateToken   func() (string, error)
	hashToken       func(token string) string
	now             func() time.Time
	logger          yca_log.Logger
	auditLogService audit_log_service.Service
}

func New(deps *Dependencies) Service {
	return &service{
		validator:       deps.Validator,
		repos:           deps.Repos,
		authorizer:      deps.Authorizer,
		generateID:      deps.GenerateID,
		generateToken:   deps.GenerateToken,
		hashToken:       deps.HashToken,
		now:             deps.Now,
		logger:          deps.Logger,
		auditLogService: deps.AuditLogService,
	}
}

var assignablePermissions = []string{
	constants.PERMISSION_ORG_READ,
	constants.PERMISSION_MEMBERS_READ,
	constants.PERMISSION_AUDIT_READ,
	constants.PERMISSION_SUBSCRIPTION_READ,
	constants.PERMISSION_ROLE_READ,
	constants.PERMISSION_TEAM_READ,
	constants.PERMISSION_TEAM_MEMBER_READ,
	constants.PERMISSION_API_KEY_READ,
}
