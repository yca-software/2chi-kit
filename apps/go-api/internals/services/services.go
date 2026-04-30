package services

import (
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	api_key_service "github.com/yca-software/2chi-kit/go-api/internals/services/api_key"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	auth_service "github.com/yca-software/2chi-kit/go-api/internals/services/auth"
	google_service "github.com/yca-software/2chi-kit/go-api/internals/services/google"
	invitation_service "github.com/yca-software/2chi-kit/go-api/internals/services/invitation"
	organization_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization"
	organization_member_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization_member"
	paddle_service "github.com/yca-software/2chi-kit/go-api/internals/services/paddle"
	role_service "github.com/yca-software/2chi-kit/go-api/internals/services/role"
	support_service "github.com/yca-software/2chi-kit/go-api/internals/services/support"
	team_service "github.com/yca-software/2chi-kit/go-api/internals/services/team"
	team_member_service "github.com/yca-software/2chi-kit/go-api/internals/services/team_member"
	user_service "github.com/yca-software/2chi-kit/go-api/internals/services/user"
	user_refresh_token_service "github.com/yca-software/2chi-kit/go-api/internals/services/user_refresh_token"
	yca_email "github.com/yca-software/go-common/email"
	yca_translate "github.com/yca-software/go-common/localizer"
	yca_log "github.com/yca-software/go-common/logger"
	yca_password "github.com/yca-software/go-common/password"
	yca_token "github.com/yca-software/go-common/token"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Services struct {
	ApiKey             api_key_service.Service
	AuditLog           audit_log_service.Service
	Auth               auth_service.Service
	Google             google_service.Service
	Invitation         invitation_service.Service
	Organization       organization_service.Service
	OrganizationMember organization_member_service.Service
	Paddle             paddle_service.Service
	Role               role_service.Service
	Support            support_service.Service
	Team               team_service.Service
	TeamMember         team_member_service.Service
	User               user_service.Service
	UserRefreshToken   user_refresh_token_service.Service
}

func NewServices(
	repos *repositories.Repositories,
	validator yca_validate.Validator,
	translator yca_translate.Translator,
	logger yca_log.Logger,
	rootPath string,
) (*Services, error) {
	emailService := yca_email.NewEmailService(&yca_email.Config{
		ResendAPIKey:  os.Getenv("RESEND_API_KEY"),
		FromEmail:     os.Getenv("EMAIL_FROM_EMAIL"),
		FromName:      os.Getenv("EMAIL_FROM_NAME"),
		TemplatesPath: rootPath + "/templates",
	})

	paddleClient, err := paddle_service.NewClient(os.Getenv("PADDLE_API_KEY"), os.Getenv("PADDLE_ENVIRONMENT"))
	if err != nil {
		logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.NewServices",
			Error:    err,
			Message:  "failed to create paddle client",
		})
		return nil, err
	}

	authorizer := helpers.NewAuthorizer(time.Now)

	auditLogService := audit_log_service.New(&audit_log_service.Dependencies{
		Validator:  validator,
		Repos:      repos,
		Authorizer: authorizer,
		GenerateID: uuid.NewV7,
		Now:        time.Now,
		Logger:     logger,
	})

	googleService := google_service.NewService(&google_service.Dependencies{
		OAuthClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		OAuthClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		OAuthRedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"),
		MapsAPIKey:        os.Getenv("GOOGLE_MAPS_API_KEY"),
		Logger:            logger,
		HTTPClient:        &http.Client{Timeout: 10 * time.Second},
	})

	paddleService := paddle_service.NewService(&paddle_service.Dependencies{
		Validator:       validator,
		Logger:          logger,
		Repos:           repos,
		Authorizer:      authorizer,
		PaddleClient:    paddleClient,
		AuditLogService: auditLogService,
		PriceIDs: &paddle_service.PriceIDs{
			BasicMonthly: os.Getenv("PADDLE_PRICE_BASIC_MONTHLY"),
			BasicAnnual:  os.Getenv("PADDLE_PRICE_BASIC_ANNUAL"),
			ProMonthly:   os.Getenv("PADDLE_PRICE_PRO_MONTHLY"),
			ProAnnual:    os.Getenv("PADDLE_PRICE_PRO_ANNUAL"),
		},
	})

	invitationService := invitation_service.New(&invitation_service.Dependencies{
		InvitationTTL:  os.Getenv("TOKEN_INVITATION_TTL"),
		ApplicationURL: os.Getenv("WEB_APP_URL"),
		Validator:      validator,
		Repos:          repos,
		Authorizer:     authorizer,
		GenerateID:     uuid.NewV7,
		Now:            time.Now,
		Logger:         logger,
		Translator:     translator,
		EmailService:   emailService,
		GenerateToken:  yca_token.GenerateToken,
		HashToken:      yca_token.HashToken,
	})

	return &Services{
		ApiKey: api_key_service.New(&api_key_service.Dependencies{
			Validator:       validator,
			Repos:           repos,
			Authorizer:      authorizer,
			GenerateID:      uuid.NewV7,
			GenerateToken:   yca_token.GenerateToken,
			HashToken:       yca_token.HashToken,
			Now:             time.Now,
			Logger:          logger,
			AuditLogService: auditLogService,
		}),
		AuditLog: auditLogService,
		Auth: auth_service.New(&auth_service.Dependencies{
			AccessTTL:            os.Getenv("TOKEN_ACCESS_TTL"),
			RefreshTTL:           os.Getenv("TOKEN_REFRESH_TTL"),
			EmailVerificationTTL: os.Getenv("TOKEN_EMAIL_VERIFICATION_TTL"),
			AccessSecret:         os.Getenv("TOKEN_ACCESS_SECRET"),
			AppURL:               os.Getenv("WEB_APP_URL"),
			Validator:            validator,
			Repos:                repos,
			Authorizer:           authorizer,
			GenerateID:           uuid.NewV7,
			GenerateToken:        yca_token.GenerateToken,
			HashToken:            yca_token.HashToken,
			Now:                  time.Now,
			Logger:               logger,
			AuditLogService:      auditLogService,
			GoogleService:        googleService,
			EmailService:         emailService,
			Translator:           translator,
			PasswordHashFn:       func(p string) (string, error) { return yca_password.Hash(p) },
			PasswordCompareFn:    yca_password.Compare,
		}),
		Google:     googleService,
		Invitation: invitationService,
		Organization: organization_service.NewService(&organization_service.Dependencies{
			Validator:         validator,
			Logger:            logger,
			Repos:             repos,
			Authorizer:        authorizer,
			PaddleService:     paddleService,
			GenerateID:        uuid.NewV7,
			Now:               time.Now,
			AuditLogService:   auditLogService,
			GoogleService:     googleService,
			InvitationService: invitationService,
		}),
		OrganizationMember: organization_member_service.NewService(&organization_member_service.Dependencies{
			Validator:       validator,
			Repositories:    repos,
			Authorizer:      authorizer,
			AuditLogService: auditLogService,
			Logger:          logger,
		}),
		Paddle: paddleService,
		Role: role_service.NewService(&role_service.Dependencies{
			GenerateID:      uuid.NewV7,
			Now:             time.Now,
			Validator:       validator,
			Repositories:    repos,
			Authorizer:      authorizer,
			AuditLogService: auditLogService,
			Logger:          logger,
		}),
		Support: support_service.New(&support_service.Dependencies{
			EmailService:      emailService,
			Now:               time.Now,
			Validator:         validator,
			SupportInboxEmail: os.Getenv("SUPPORT_INBOX_EMAIL"),
		}),
		Team: team_service.NewService(&team_service.Dependencies{
			GenerateID:      uuid.NewV7,
			Now:             time.Now,
			Validator:       validator,
			Repositories:    repos,
			Authorizer:      authorizer,
			AuditLogService: auditLogService,
			Logger:          logger,
		}),
		TeamMember: team_member_service.NewService(&team_member_service.Dependencies{
			GenerateID:      uuid.NewV7,
			Now:             time.Now,
			Validator:       validator,
			Repositories:    repos,
			Authorizer:      authorizer,
			AuditLogService: auditLogService,
			Logger:          logger,
		}),
		User: user_service.NewService(&user_service.Dependencies{
			GenerateID:     uuid.NewV7,
			Now:            time.Now,
			Validator:      validator,
			Repositories:   repos,
			Authorizer:     authorizer,
			Logger:         logger,
			PasswordHashFn: func(p string) (string, error) { return yca_password.Hash(p) },
			GenerateToken:  yca_token.GenerateToken,
			HashToken:      yca_token.HashToken,
		}),
		UserRefreshToken: user_refresh_token_service.NewService(&user_refresh_token_service.Dependencies{
			Now:          time.Now,
			Validator:    validator,
			Repositories: repos,
			Authorizer:   authorizer,
			Logger:       logger,
			HashToken:    yca_token.HashToken,
		}),
	}, nil
}
