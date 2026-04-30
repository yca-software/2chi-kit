package organization_repository

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_observe "github.com/yca-software/go-common/observer"
	yca_repository "github.com/yca-software/go-common/repository"
)

var (
	TABLE_NAME = "organizations"
	COLUMNS    = []string{"id", "created_at", "deleted_at", "name", "address", "city", "zip", "country", "place_id", "geo", "timezone", "billing_email", "custom_subscription", "subscription_expires_at", "subscription_payment_interval", "subscription_type", "subscription_seats", "subscription_in_trial", "paddle_subscription_id", "paddle_customer_id", "scheduled_plan_price_id"}
)

type Repository interface {
	yca_repository.Repository[models.Organization]
	Archive(tx yca_repository.Tx, id string) error
	CleanupArchived() error
	Count() (int, error)
	Create(tx yca_repository.Tx, org *models.Organization) error
	Delete(tx yca_repository.Tx, id string) error
	GetByID(id string) (*models.Organization, error)
	GetByIDIncludeArchived(id string) (*models.Organization, error)
	GetByPaddleCustomerID(paddleCustomerID string) (*models.Organization, error)
	GetOrganizationsWithScheduledPlanChangeDue() (*[]models.Organization, error)
	Search(searchPhrase string, limit, offset int) (*[]models.Organization, error)
	SearchArchived(searchPhrase string, limit, offset int) (*[]models.Organization, error)
	Restore(tx yca_repository.Tx, id string) error
	Update(tx yca_repository.Tx, org *models.Organization) error
}

type repository struct {
	yca_repository.Repository[models.Organization]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.Organization](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Archive(tx yca_repository.Tx, id string) error {
	return r.BaseUpdate(tx, squirrel.Eq{"id": id, "deleted_at": nil}, map[string]any{
		"deleted_at": time.Now(),
	})
}

func (r *repository) CleanupArchived() error {
	threshold := time.Now().Add(-constants.StaleDataRetentionPeriod)
	return r.BaseDelete(nil, squirrel.And{
		squirrel.LtOrEq{"deleted_at": threshold},
	})
}

func (r *repository) Count() (int, error) {
	return r.BaseCount(nil, squirrel.Eq{"deleted_at": nil})
}

func (r *repository) Create(tx yca_repository.Tx, org *models.Organization) error {
	return r.BaseCreate(tx, map[string]any{
		"id":                            org.ID,
		"created_at":                    org.CreatedAt,
		"name":                          org.Name,
		"address":                       org.Address,
		"city":                          org.City,
		"zip":                           org.Zip,
		"country":                       org.Country,
		"place_id":                      org.PlaceID,
		"geo":                           org.Geo,
		"timezone":                      org.Timezone,
		"billing_email":                 org.BillingEmail,
		"custom_subscription":           org.CustomSubscription,
		"subscription_expires_at":       org.SubscriptionExpiresAt,
		"subscription_payment_interval": org.SubscriptionPaymentInterval,
		"subscription_type":             org.SubscriptionType,
		"subscription_seats":            org.SubscriptionSeats,
		"subscription_in_trial":         org.SubscriptionInTrial,
		"paddle_customer_id":            org.PaddleCustomerID,
		"scheduled_plan_price_id":       org.ScheduledPlanPriceID,
	})
}

func (r *repository) Delete(tx yca_repository.Tx, id string) error {
	return r.BaseDelete(tx, squirrel.Eq{"id": id})
}

func (r *repository) GetByID(id string) (*models.Organization, error) {
	return r.BaseGet(nil, squirrel.Eq{"id": id, "deleted_at": nil}, nil)
}

func (r *repository) GetByIDIncludeArchived(id string) (*models.Organization, error) {
	return r.BaseGet(nil, squirrel.Eq{"id": id}, nil)
}

func (r *repository) GetByPaddleCustomerID(paddleCustomerID string) (*models.Organization, error) {
	return r.BaseGet(nil, squirrel.Eq{"paddle_customer_id": paddleCustomerID, "deleted_at": nil}, nil)
}

func (r *repository) GetOrganizationsWithScheduledPlanChangeDue() (*[]models.Organization, error) {
	condition := squirrel.And{
		squirrel.Eq{"deleted_at": nil},
		squirrel.Expr("scheduled_plan_price_id IS NOT NULL"),
		squirrel.LtOrEq{"subscription_expires_at": time.Now()},
	}
	return r.BaseSelect(nil, condition, nil, "subscription_expires_at ASC")
}

func (r *repository) Search(searchPhrase string, limit, offset int) (*[]models.Organization, error) {
	condition := squirrel.And{
		squirrel.Eq{"deleted_at": nil},
	}
	if searchPhrase != "" {
		condition = append(condition, squirrel.Or{
			squirrel.ILike{"name": "%" + searchPhrase + "%"},
			squirrel.ILike{"address": "%" + searchPhrase + "%"},
		})
	}
	return r.BasePaginatedSelect(nil, condition, nil, "created_at DESC", uint64(limit), uint64(offset))
}

func (r *repository) SearchArchived(searchPhrase string, limit, offset int) (*[]models.Organization, error) {
	condition := squirrel.And{
		squirrel.Expr("deleted_at IS NOT NULL"),
	}
	if searchPhrase != "" {
		condition = append(condition, squirrel.Or{
			squirrel.ILike{"name": "%" + searchPhrase + "%"},
			squirrel.ILike{"address": "%" + searchPhrase + "%"},
		})
	}
	return r.BasePaginatedSelect(nil, condition, nil, "deleted_at DESC", uint64(limit), uint64(offset))
}

func (r *repository) Restore(tx yca_repository.Tx, id string) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Expr("deleted_at IS NOT NULL"),
	}, map[string]any{
		"deleted_at": nil,
	})
}

func (r *repository) Update(tx yca_repository.Tx, org *models.Organization) error {
	return r.BaseUpdate(tx, squirrel.Eq{"id": org.ID, "deleted_at": nil}, map[string]any{
		"name":                          org.Name,
		"address":                       org.Address,
		"city":                          org.City,
		"zip":                           org.Zip,
		"country":                       org.Country,
		"place_id":                      org.PlaceID,
		"geo":                           org.Geo,
		"timezone":                      org.Timezone,
		"billing_email":                 org.BillingEmail,
		"custom_subscription":           org.CustomSubscription,
		"subscription_expires_at":       org.SubscriptionExpiresAt,
		"subscription_payment_interval": org.SubscriptionPaymentInterval,
		"subscription_type":             org.SubscriptionType,
		"subscription_seats":            org.SubscriptionSeats,
		"subscription_in_trial":         org.SubscriptionInTrial,
		"paddle_subscription_id":        org.PaddleSubscriptionID,
		"scheduled_plan_price_id":       org.ScheduledPlanPriceID,
	})
}
