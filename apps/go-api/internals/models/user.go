package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	FirstName string `json:"firstName" db:"first_name"`
	LastName  string `json:"lastName" db:"last_name"`
	Language  string `json:"language" db:"language"`
	AvatarURL string `json:"avatarURL" db:"avatar_url"`

	Email           string     `json:"email" db:"email"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt" db:"email_verified_at"`
	Password        *string    `json:"password" db:"password"`
	GoogleID        *string    `json:"googleId" db:"google_id"`

	TermsAcceptedAt time.Time `json:"termsAcceptedAt" db:"terms_accepted_at"`
	TermsVersion    string    `json:"termsVersion" db:"terms_version"`
}

type AdminAccess struct {
	UserID    string    `json:"userId" db:"user_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type UserRefreshToken struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"userId" db:"user_id"`

	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	ExpiresAt time.Time  `json:"expiresAt" db:"expires_at"`
	RevokedAt *time.Time `json:"revokedAt" db:"revoked_at"`

	IP             string        `json:"ip" db:"ip"`
	UserAgent      string        `json:"userAgent" db:"user_agent"`
	TokenHash      string        `json:"-" db:"token_hash"`
	ImpersonatedBy uuid.NullUUID `json:"-" db:"impersonated_by"`
}

type UserPasswordResetToken struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"userId" db:"user_id"`

	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	ExpiresAt time.Time  `json:"expiresAt" db:"expires_at"`
	UsedAt    *time.Time `json:"usedAt" db:"used_at"`

	TokenHash string `json:"-" db:"token_hash"`
}

type UserEmailVerificationToken struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"userId" db:"user_id"`

	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	ExpiresAt time.Time  `json:"expiresAt" db:"expires_at"`
	UsedAt    *time.Time `json:"usedAt" db:"used_at"`

	TokenHash string `json:"-" db:"token_hash"`
}
