package organization_service

import (
	"time"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	google_service "github.com/yca-software/2chi-kit/go-api/internals/services/google"
	invitation_service "github.com/yca-software/2chi-kit/go-api/internals/services/invitation"
	paddle_service "github.com/yca-software/2chi-kit/go-api/internals/services/paddle"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	Validator         yca_validate.Validator
	Logger            yca_log.Logger
	Repos             *repositories.Repositories
	Authorizer        *helpers.Authorizer
	AuditLogService   audit_log_service.Service
	PaddleService     paddle_service.Service
	GoogleService     google_service.Service
	InvitationService invitation_service.Service
	GenerateID        func() (uuid.UUID, error)
	Now               func() time.Time
}

type Service interface {
	AdminCreateOrganizationWithCustomSubscription(req *AdminCreateOrganizationWithCustomSubscriptionRequest, accessInfo *models.AccessInfo) (*models.Organization, error)
	AdminUpdateSubscriptionSettings(req *AdminUpdateSubscriptionSettingsRequest, accessInfo *models.AccessInfo) (*models.Organization, error)
	Archive(req *ArchiveRequest, accessInfo *models.AccessInfo) error
	CleanupArchived() error
	Count(accessInfo *models.AccessInfo) (int, error)
	Create(req *CreateRequest, accessInfo *models.AccessInfo) (*CreateResponse, error)
	Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error
	Get(req *GetRequest, accessInfo *models.AccessInfo) (*models.Organization, error)
	GetArchived(req *GetRequest, accessInfo *models.AccessInfo) (*models.Organization, error)
	List(req *ListRequest, accessInfo *models.AccessInfo) (*PaginatedListResponse, error)
	ListArchived(req *ListRequest, accessInfo *models.AccessInfo) (*PaginatedListResponse, error)
	Restore(req *RestoreRequest, accessInfo *models.AccessInfo) error
	Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.Organization, error)
}

type service struct {
	validator         yca_validate.Validator
	logger            yca_log.Logger
	repos             *repositories.Repositories
	authorizer        *helpers.Authorizer
	auditLogService   audit_log_service.Service
	paddleService     paddle_service.Service
	googleService     google_service.Service
	invitationService invitation_service.Service
	generateID        func() (uuid.UUID, error)
	now               func() time.Time
}

func NewService(deps *Dependencies) Service {
	return &service{
		validator:         deps.Validator,
		logger:            deps.Logger,
		repos:             deps.Repos,
		authorizer:        deps.Authorizer,
		auditLogService:   deps.AuditLogService,
		paddleService:     deps.PaddleService,
		googleService:     deps.GoogleService,
		invitationService: deps.InvitationService,
		generateID:        deps.GenerateID,
		now:               deps.Now,
	}
}
