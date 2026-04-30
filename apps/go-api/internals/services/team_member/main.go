package team_member_service

import (
	"time"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	GenerateID      func() (uuid.UUID, error)
	Now             func() time.Time
	Validator       yca_validate.Validator
	Repositories    *repositories.Repositories
	Authorizer      *helpers.Authorizer
	Logger          yca_log.Logger
	AuditLogService audit_log_service.Service
}

type Service interface {
	Add(req *AddRequest, accessInfo *models.AccessInfo) (*models.TeamMemberWithUser, error)
	ListByTeam(req *ListByTeamRequest, accessInfo *models.AccessInfo) (*[]models.TeamMemberWithUser, error)
	Remove(req *RemoveRequest, accessInfo *models.AccessInfo) error
}

type service struct {
	generateID      func() (uuid.UUID, error)
	now             func() time.Time
	validator       yca_validate.Validator
	repos           *repositories.Repositories
	authorizer      *helpers.Authorizer
	logger          yca_log.Logger
	auditLogService audit_log_service.Service
}

func NewService(deps *Dependencies) Service {
	return &service{
		generateID:      deps.GenerateID,
		now:             deps.Now,
		validator:       deps.Validator,
		repos:           deps.Repositories,
		authorizer:      deps.Authorizer,
		logger:          deps.Logger,
		auditLogService: deps.AuditLogService,
	}
}
