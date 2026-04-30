package organization_member_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	Validator       yca_validate.Validator
	Repositories    *repositories.Repositories
	Authorizer      *helpers.Authorizer
	Logger          yca_log.Logger
	AuditLogService audit_log_service.Service
}

type Service interface {
	ListByOrganization(req *ListByOrganizationRequest, accessInfo *models.AccessInfo) (*[]models.OrganizationMemberWithUser, error)
	ListByUser(req *ListByUserRequest, accessInfo *models.AccessInfo) (*[]models.OrganizationMemberWithOrganization, error)
	Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.OrganizationMemberWithUser, error)
	Remove(req *RemoveRequest, accessInfo *models.AccessInfo) error
}

type service struct {
	validator       yca_validate.Validator
	repos           *repositories.Repositories
	authorizer      *helpers.Authorizer
	logger          yca_log.Logger
	auditLogService audit_log_service.Service
}

func NewService(deps *Dependencies) Service {
	return &service{
		validator:       deps.Validator,
		repos:           deps.Repositories,
		authorizer:      deps.Authorizer,
		logger:          deps.Logger,
		auditLogService: deps.AuditLogService,
	}
}
