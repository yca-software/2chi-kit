package support_service

import (
	"strings"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_email "github.com/yca-software/go-common/email"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	SupportInboxEmail string
	EmailService      yca_email.EmailService
	Now               func() time.Time
	Validator         yca_validate.Validator
}

type Service interface {
	Submit(req *SubmitRequest, accessInfo *models.AccessInfo) error
}

type service struct {
	supportInboxEmail string
	emailService      yca_email.EmailService
	now               func() time.Time
	validator         yca_validate.Validator
}

func New(deps *Dependencies) Service {
	return &service{
		supportInboxEmail: strings.TrimSpace(deps.SupportInboxEmail),
		emailService:      deps.EmailService,
		now:               deps.Now,
		validator:         deps.Validator,
	}
}
