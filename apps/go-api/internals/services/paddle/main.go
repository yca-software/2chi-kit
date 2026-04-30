package paddle_service

import (
	"context"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

// PriceIDs are Paddle price IDs for this app (e.g. from PADDLE_PRICE_BASIC_MONTHLY, PADDLE_PRICE_BASIC_ANNUAL,
// PADDLE_PRICE_PRO_MONTHLY, PADDLE_PRICE_PRO_ANNUAL). Webhooks and checkout use these to recognize this product’s subscriptions.
type PriceIDs struct {
	BasicMonthly string
	BasicAnnual  string
	ProMonthly   string
	ProAnnual    string
}

type Dependencies struct {
	Validator       yca_validate.Validator
	Logger          yca_log.Logger
	Repos           *repositories.Repositories
	Authorizer      *helpers.Authorizer
	PaddleClient    PaddleClient
	PriceIDs        *PriceIDs // set from env in production; nil or empty fields mean no price is recognized
	AuditLogService audit_log_service.Service
}

type Service interface {
	CancelSubscription(paddleSubscriptionID string) error
	CreateCheckoutSession(req CreateCheckoutSessionRequest, accessInfo *models.AccessInfo) (*CheckoutSessionResponse, error)
	CreateCustomer(organization *models.Organization) (*paddle.Customer, error)
	CreateCustomerPortalSession(req CreateCustomerPortalSessionRequest, accessInfo *models.AccessInfo) (*CustomerPortalSessionResponse, error)
	HandleWebhook(payload []byte, signature string) error
	UpdateCustomer(organization *models.Organization) (*paddle.Customer, error)
	ProcessTransaction(ctx context.Context, req *ProcessTransactionRequest, accessInfo *models.AccessInfo) (*models.Organization, error)
	ChangePlan(ctx context.Context, req *ChangePlanRequest, accessInfo *models.AccessInfo) (*ChangePlanResult, error)
	ApplyScheduledPlanChanges(ctx context.Context) error
}

type service struct {
	validator       yca_validate.Validator
	logger          yca_log.Logger
	repos           *repositories.Repositories
	authorizer      *helpers.Authorizer
	paddleClient    PaddleClient
	priceIDs        *PriceIDs
	auditLogService audit_log_service.Service
}

func NewService(deps *Dependencies) Service {
	return &service{
		validator:       deps.Validator,
		logger:          deps.Logger,
		repos:           deps.Repos,
		authorizer:      deps.Authorizer,
		paddleClient:    deps.PaddleClient,
		priceIDs:        deps.PriceIDs,
		auditLogService: deps.AuditLogService,
	}
}

func (s *service) paddlePriceIDs() (basicMonthly, basicAnnual, proMonthly, proAnnual string) {
	if s.priceIDs == nil {
		return "", "", "", ""
	}
	return s.priceIDs.BasicMonthly, s.priceIDs.BasicAnnual, s.priceIDs.ProMonthly, s.priceIDs.ProAnnual
}

// isOurPriceID returns true only when the price ID is one of this product's prices.
// A customer with a subscription for another product's pricing does not count as having a subscription here.
func (s *service) isOurPriceID(priceID string) bool {
	if priceID == "" {
		return false
	}
	basicMonthly, basicAnnual, proMonthly, proAnnual := s.paddlePriceIDs()
	return priceID == basicMonthly || priceID == basicAnnual || priceID == proMonthly || priceID == proAnnual
}

func (s *service) getTierFromPriceID(priceID string) int {
	basicMonthly, basicAnnual, proMonthly, proAnnual := s.paddlePriceIDs()
	if priceID == basicMonthly || priceID == basicAnnual {
		return constants.SUBSCRIPTION_TYPE_BASIC
	}
	if priceID == proMonthly || priceID == proAnnual {
		return constants.SUBSCRIPTION_TYPE_PRO
	}
	return constants.SUBSCRIPTION_TYPE_BASIC
}

func getSeatsByTier(tier int) int {
	switch tier {
	case constants.SUBSCRIPTION_TYPE_BASIC:
		return constants.SUBSCRIPTION_TYPE_SEATS_INCLUDED_BASIC
	case constants.SUBSCRIPTION_TYPE_PRO:
		return constants.SUBSCRIPTION_TYPE_SEATS_INCLUDED_PRO
	case constants.SUBSCRIPTION_TYPE_ENTERPRISE:
		return constants.SUBSCRIPTION_TYPE_SEATS_INCLUDED_ENTERPRISE
	default:
		return constants.SUBSCRIPTION_TYPE_SEATS_INCLUDED_FREE
	}
}

func (s *service) getIntervalFromPriceID(priceID string) int {
	_, basicAnnual, _, proAnnual := s.paddlePriceIDs()
	if priceID != "" && (priceID == basicAnnual || priceID == proAnnual) {
		return constants.PAYMENT_INTERVAL_ANNUAL
	}
	return constants.PAYMENT_INTERVAL_MONTHLY
}

// applySubscriptionFromPrice mutates the given organization in-memory using
// the subscription tier and interval derived from the provided Paddle price ID.
// The caller is responsible for persisting the organization.
func (s *service) applySubscriptionFromPrice(org *models.Organization, priceID string) {
	tier := s.getTierFromPriceID(priceID)
	org.SubscriptionType = tier
	org.SubscriptionSeats = getSeatsByTier(tier)
	org.SubscriptionPaymentInterval = s.getIntervalFromPriceID(priceID)
}
