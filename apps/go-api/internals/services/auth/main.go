package auth_service

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	google_service "github.com/yca-software/2chi-kit/go-api/internals/services/google"
	email_service "github.com/yca-software/go-common/email"
	yca_error "github.com/yca-software/go-common/error"
	yca_translate "github.com/yca-software/go-common/localizer"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type Dependencies struct {
	AccessSecret         string
	AccessTTL            string
	RefreshTTL           string
	EmailVerificationTTL string
	PasswordResetTTL     string
	AppURL               string
	Validator            yca_validate.Validator
	Repos                *repositories.Repositories
	Authorizer           *helpers.Authorizer
	GenerateID           func() (uuid.UUID, error)
	GenerateToken        func() (string, error)
	HashToken            func(token string) string
	Now                  func() time.Time
	Logger               yca_log.Logger
	AuditLogService      audit_log_service.Service
	GoogleService        google_service.Service
	EmailService         email_service.EmailService
	Translator           yca_translate.Translator
	PasswordHashFn       func(password string) (string, error)
	PasswordCompareFn    func(password, encodedHash string) bool
}

type Service interface {
	AuthenticateWithGoogle(req *AuthenticateWithGoogleRequest) (*AuthenticateResponse, error)
	AuthenticateWithPassword(req *AuthenticateWithPasswordRequest) (*AuthenticateResponse, error)
	ForgotPassword(req *ForgotPasswordRequest) error
	Logout(req *LogoutRequest, accessInfo *models.AccessInfo) error
	RefreshAccessToken(req *RefreshAccessTokenRequest) (*RefreshAccessTokenResponse, error)
	ResetPassword(req *ResetPasswordRequest) error
	SignUp(req *SignUpRequest) (*SignUpResponse, error)
	VerifyEmail(req *VerifyEmailRequest) error
	ResendVerificationEmail(req *ResendVerificationEmailRequest) error
	Impersonate(req *ImpersonateRequest, accessInfo *models.AccessInfo) (*AuthenticateResponse, error)
	CleanupStalePasswordResetTokens() error
	CleanupStaleEmailVerificationTokens() error
}

type service struct {
	refreshTTL           int
	accessTTL            int
	emailVerificationTTL int
	passwordResetTTL     int
	accessSecret         string
	appURL               string
	validator            yca_validate.Validator
	repos                *repositories.Repositories
	authorizer           *helpers.Authorizer
	generateID           func() (uuid.UUID, error)
	generateToken        func() (string, error)
	hashToken            func(token string) string
	now                  func() time.Time
	logger               yca_log.Logger
	translator           yca_translate.Translator
	auditLogService      audit_log_service.Service
	googleService        google_service.Service
	emailService         email_service.EmailService
	passwordHashFn       func(password string) (string, error)
	passwordCompareFn    func(password, encodedHash string) bool
}

func New(deps *Dependencies) Service {
	accessTTL, err := strconv.Atoi(deps.AccessTTL)
	if err != nil {
		accessTTL = 15
	}
	refreshTTL, err := strconv.Atoi(deps.RefreshTTL)
	if err != nil {
		refreshTTL = 168
	}
	emailVerificationTTL, err := strconv.Atoi(deps.EmailVerificationTTL)
	if err != nil {
		emailVerificationTTL = 168
	}
	passwordResetTTL, err := strconv.Atoi(deps.PasswordResetTTL)
	if err != nil {
		passwordResetTTL = 24
	}
	return &service{
		accessTTL:            accessTTL,
		refreshTTL:           refreshTTL,
		emailVerificationTTL: emailVerificationTTL,
		passwordResetTTL:     passwordResetTTL,
		accessSecret:         deps.AccessSecret,
		appURL:               deps.AppURL,
		validator:            deps.Validator,
		repos:                deps.Repos,
		authorizer:           deps.Authorizer,
		generateID:           deps.GenerateID,
		generateToken:        deps.GenerateToken,
		hashToken:            deps.HashToken,
		now:                  deps.Now,
		logger:               deps.Logger,
		auditLogService:      deps.AuditLogService,
		googleService:        deps.GoogleService,
		emailService:         deps.EmailService,
		translator:           deps.Translator,
		passwordHashFn:       deps.PasswordHashFn,
		passwordCompareFn:    deps.PasswordCompareFn,
	}
}

type AuthenticateResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (s *service) generateAccessToken(user *models.User, impersonatedBy string, impersonatedByEmail string) (string, error) {
	adminAccess, err := s.repos.AdminAccess.GetByUserID(user.ID.String())
	if err != nil {
		if e, ok := err.(*yca_error.Error); ok {
			if e.ErrorCode != constants.NOT_FOUND_CODE {
				return "", e
			}
		} else {
			return "", err
		}
	}

	isAdmin := adminAccess != nil

	roles, err := s.repos.OrganizationMember.ListByUserIDWithRole(user.ID.String())
	if err != nil {
		return "", err
	}

	permissions := make([]models.JWTAccessTokenPermissionData, len(*roles))
	for i, role := range *roles {
		permissions[i] = models.JWTAccessTokenPermissionData{
			OrganizationID: role.OrganizationID,
			RoleID:         role.RoleID,
			Permissions:    role.RolePermissions,
		}
	}

	claims := jwt.MapClaims{
		"sub":         user.ID.String(),
		"email":       user.Email,
		"exp":         s.now().Add(time.Duration(s.accessTTL) * time.Minute).Unix(),
		"iat":         s.now().Unix(),
		"permissions": permissions,
		"isAdmin":     isAdmin,
	}
	if impersonatedBy != "" {
		claims["impersonatedBy"] = impersonatedBy
		claims["impersonatedByEmail"] = impersonatedByEmail
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.accessSecret))
}
