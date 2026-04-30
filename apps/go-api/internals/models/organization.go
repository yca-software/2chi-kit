package models

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	DeletedAt *time.Time `db:"deleted_at" json:"deletedAt"`

	Name string `db:"name" json:"name"`

	Address  string `db:"address" json:"address"`
	City     string `db:"city" json:"city"`
	Zip      string `db:"zip" json:"zip"`
	Country  string `db:"country" json:"country"`
	PlaceID  string `db:"place_id" json:"placeId"`
	Geo      Point  `db:"geo" json:"geo"`
	Timezone string `db:"timezone" json:"timezone"`

	BillingEmail                string     `db:"billing_email" json:"billingEmail"`
	CustomSubscription          bool       `db:"custom_subscription" json:"customSubscription"`
	SubscriptionExpiresAt       *time.Time `db:"subscription_expires_at" json:"subscriptionExpiresAt"`
	SubscriptionPaymentInterval int        `db:"subscription_payment_interval" json:"subscriptionPaymentInterval"`
	SubscriptionType            int        `db:"subscription_type" json:"subscriptionType"`
	SubscriptionSeats           int        `db:"subscription_seats" json:"subscriptionSeats"`
	SubscriptionInTrial         bool       `db:"subscription_in_trial" json:"subscriptionInTrial"`

	PaddleSubscriptionID *string `db:"paddle_subscription_id" json:"paddleSubscriptionId"`
	PaddleCustomerID     string  `db:"paddle_customer_id" json:"paddleCustomerId"`

	// When set: switch to this Paddle price at end of current period (subscription_expires_at). Used for annual→monthly.
	ScheduledPlanPriceID *string `db:"scheduled_plan_price_id" json:"scheduledPlanPriceId"`
}

type Role struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`

	OrganizationID uuid.UUID `db:"organization_id" json:"organizationId"`

	Name        string          `db:"name" json:"name"`
	Description string          `db:"description" json:"description"`
	Permissions RolePermissions `db:"permissions" json:"permissions"`
	Locked      bool            `db:"locked" json:"locked"`
}

type RolePermissions []string

func (rp RolePermissions) Value() (driver.Value, error) {
	return JSONBValue(rp)
}

func (rp *RolePermissions) Scan(value any) error {
	return JSONBScan(value, rp)
}

type OrganizationMember struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`

	OrganizationID uuid.UUID `db:"organization_id" json:"organizationId"`
	UserID         uuid.UUID `db:"user_id" json:"userId"`
	RoleID         uuid.UUID `db:"role_id" json:"roleId"`
}

type OrganizationMemberWithOrganization struct {
	OrganizationMember
	OrganizationName string `db:"organization_name" json:"organizationName"`
}

type OrganizationMemberWithOrganizationAndRole struct {
	OrganizationMemberWithOrganization
	RoleName        string          `db:"role_name" json:"roleName"`
	RolePermissions RolePermissions `db:"role_permissions" json:"rolePermissions"`
}

type OrganizationMemberWithUser struct {
	OrganizationMember
	UserEmail     string `db:"user_email" json:"userEmail"`
	UserFirstName string `db:"user_first_name" json:"userFirstName"`
	UserLastName  string `db:"user_last_name" json:"userLastName"`
}

type Team struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`

	OrganizationID uuid.UUID `db:"organization_id" json:"organizationId"`

	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

type TeamMember struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`

	OrganizationID uuid.UUID `db:"organization_id" json:"organizationId"`
	TeamID         uuid.UUID `db:"team_id" json:"teamId"`
	UserID         uuid.UUID `db:"user_id" json:"userId"`
}

type TeamMemberWithTeam struct {
	TeamMember
	TeamName string `db:"team_name" json:"teamName"`
}

type TeamMemberWithUser struct {
	TeamMember
	UserEmail     string `db:"user_email" json:"userEmail"`
	UserFirstName string `db:"user_first_name" json:"userFirstName"`
	UserLastName  string `db:"user_last_name" json:"userLastName"`
}

type Invitation struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	ExpiresAt  time.Time  `db:"expires_at" json:"expiresAt"`
	AcceptedAt *time.Time `db:"accepted_at" json:"acceptedAt"`
	RevokedAt  *time.Time `db:"revoked_at" json:"revokedAt"`

	OrganizationID uuid.UUID `db:"organization_id" json:"organizationId"`
	RoleID         uuid.UUID `db:"role_id" json:"roleId"`
	Email          string    `db:"email" json:"email"`

	InvitedByID    uuid.NullUUID `db:"invited_by_id" json:"invitedById"`
	InvitedByEmail string        `db:"invited_by_email" json:"invitedByEmail"`
	TokenHash      string        `db:"token_hash" json:"-"`
}

type APIKey struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`

	Name           string          `db:"name" json:"name"`
	KeyPrefix      string          `db:"key_prefix" json:"keyPrefix"`
	KeyHash        string          `db:"key_hash" json:"-"`
	OrganizationID uuid.UUID       `db:"organization_id" json:"organizationId"`
	Permissions    RolePermissions `db:"permissions" json:"permissions"`
}
