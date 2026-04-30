package invitation_service

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	yca_email "github.com/yca-software/go-common/email"
	yca_translate "github.com/yca-software/go-common/localizer"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	InvitationTTL  string
	ApplicationURL string
	Validator      yca_validate.Validator
	Authorizer     *helpers.Authorizer
	Repos          *repositories.Repositories
	Logger         yca_log.Logger
	EmailService   yca_email.EmailService
	Now            func() time.Time
	GenerateID     func() (uuid.UUID, error)
	HashToken      func(token string) string
	GenerateToken  func() (string, error)
	Translator     yca_translate.Translator
}

type Service interface {
	Create(req *CreateRequest, accessInfo *models.AccessInfo) (*CreateResponse, error)
	Revoke(req *RevokeRequest, accessInfo *models.AccessInfo) error
	List(req *ListRequest, accessInfo *models.AccessInfo) (*[]models.Invitation, error)
	CleanupStale() error
}

type service struct {
	invitationTTL int
	appURL        string
	validator     yca_validate.Validator
	authorizer    *helpers.Authorizer
	repos         *repositories.Repositories
	logger        yca_log.Logger
	emailService  yca_email.EmailService
	now           func() time.Time
	generateID    func() (uuid.UUID, error)
	hashToken     func(token string) string
	generateToken func() (string, error)
	translator    yca_translate.Translator
}

func New(deps *Dependencies) Service {
	invitationTTL, err := strconv.Atoi(deps.InvitationTTL)
	if err != nil {
		invitationTTL = 24
	}
	return &service{
		invitationTTL: invitationTTL,
		appURL:        deps.ApplicationURL,
		validator:     deps.Validator,
		authorizer:    deps.Authorizer,
		repos:         deps.Repos,
		logger:        deps.Logger,
		emailService:  deps.EmailService,
		now:           deps.Now,
		generateID:    deps.GenerateID,
		hashToken:     deps.HashToken,
		generateToken: deps.GenerateToken,
		translator:    deps.Translator,
	}
}
