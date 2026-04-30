package repositories

import (
	"github.com/jmoiron/sqlx"
	admin_access_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/admin_access"
	api_key_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/api_key"
	audit_log_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/audit_log"
	invitation_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/invitation"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	organization_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization_member"
	role_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/role"
	team_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/team"
	team_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/team_member"
	user_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user"
	user_email_verification_token_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user_email_verification_token"
	user_password_reset_token_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user_password_reset_token"
	user_refresh_token_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user_refresh_token"
	observer "github.com/yca-software/go-common/observer"
)

type Repositories struct {
	AdminAccess                admin_access_repository.Repository
	ApiKey                     api_key_repository.Repository
	AuditLog                   audit_log_repository.Repository
	Invitation                 invitation_repository.Repository
	Organization               organization_repository.Repository
	OrganizationMember         organization_member_repository.Repository
	Role                       role_repository.Repository
	Team                       team_repository.Repository
	TeamMember                 team_member_repository.Repository
	User                       user_repository.Repository
	UserEmailVerificationToken user_email_verification_token_repository.Repository
	UserPasswordResetToken     user_password_reset_token_repository.Repository
	UserRefreshToken           user_refresh_token_repository.Repository
}

func NewRepositories(db *sqlx.DB, metricsHook observer.QueryMetricsHook) *Repositories {
	return &Repositories{
		AdminAccess:                admin_access_repository.New(db, metricsHook),
		ApiKey:                     api_key_repository.New(db, metricsHook),
		AuditLog:                   audit_log_repository.New(db, metricsHook),
		Invitation:                 invitation_repository.New(db, metricsHook),
		Organization:               organization_repository.New(db, metricsHook),
		OrganizationMember:         organization_member_repository.New(db, metricsHook),
		Role:                       role_repository.New(db, metricsHook),
		Team:                       team_repository.New(db, metricsHook),
		TeamMember:                 team_member_repository.New(db, metricsHook),
		User:                       user_repository.New(db, metricsHook),
		UserEmailVerificationToken: user_email_verification_token_repository.New(db, metricsHook),
		UserPasswordResetToken:     user_password_reset_token_repository.New(db, metricsHook),
		UserRefreshToken:           user_refresh_token_repository.New(db, metricsHook),
	}
}
