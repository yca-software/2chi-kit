package paddle_service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	paddleerr "github.com/PaddleHQ/paddle-go-sdk/v4/pkg/paddleerr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	paddle_service "github.com/yca-software/2chi-kit/go-api/internals/services/paddle"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type PaddleServiceTestSuite struct {
	suite.Suite
	svc          paddle_service.Service
	paddleClient *paddle_service.MockPaddleClient
	orgRepo      *organization_repository.MockRepository
	logger       *yca_log.MockLogger
	authorizer   *helpers.Authorizer
	repos        *repositories.Repositories
	now          time.Time
}

const testBasicPriceID = "pri_basic_monthly"

func TestPaddleServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PaddleServiceTestSuite))
}

func (s *PaddleServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	s.paddleClient = paddle_service.NewMockPaddleClient()
	s.orgRepo = organization_repository.NewMock()
	s.logger = &yca_log.MockLogger{}
	s.logger.On("Log", mock.Anything).Return().Maybe()

	s.repos = &repositories.Repositories{
		Organization: s.orgRepo,
	}
	s.authorizer = helpers.NewAuthorizer(func() time.Time { return s.now })
	auditLogSvc := audit_log_service.NewMockService()
	auditLogSvc.On("Create", mock.Anything, mock.Anything).Return(&models.AuditLog{}, nil).Maybe()

	s.svc = paddle_service.NewService(&paddle_service.Dependencies{
		Validator:       yca_validate.New(),
		Logger:          s.logger,
		Repos:           s.repos,
		Authorizer:      s.authorizer,
		PaddleClient:    s.paddleClient,
		PriceIDs:        &paddle_service.PriceIDs{BasicMonthly: testBasicPriceID, BasicAnnual: "pri_basic_yearly", ProMonthly: "pri_pro_monthly", ProAnnual: "pri_pro_yearly"},
		AuditLogService: auditLogSvc,
	})
}

// accessInfoAdmin returns AccessInfo with admin user so permission checks pass
func (s *PaddleServiceTestSuite) accessInfoAdmin(orgID string) *models.AccessInfo {
	return &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			IsAdmin: true,
			Roles:   []models.JWTAccessTokenPermissionData{{OrganizationID: uuid.MustParse(orgID), Permissions: []string{constants.PERMISSION_SUBSCRIPTION_WRITE}}},
		},
	}
}

func (s *PaddleServiceTestSuite) orgWithPaddle(orgID, paddleCustomerID string) *models.Organization {
	id := uuid.MustParse(orgID)
	return &models.Organization{
		ID:               id,
		Name:             "Test Org",
		BillingEmail:     "billing@example.com",
		PaddleCustomerID: paddleCustomerID,
		Address:          "123 Main St",
		City:             "Sydney",
		Zip:              "2000",
		Country:          "AU",
		Timezone:         "Australia/Sydney",
	}
}

// --- HandleWebhook ---

func (s *PaddleServiceTestSuite) TestHandleWebhook_InvalidJSON() {
	err := s.svc.HandleWebhook([]byte("invalid"), "sig")
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.NotEmpty(e.ErrorCode)
	}
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_UnknownEventType() {
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "unknown.event",
		Data:      map[string]any{},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.NoError(err)
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_SubscriptionUpdate_MissingSubscriptionID() {
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "subscription.created",
		Data: map[string]any{
			"subscription": map[string]any{
				"customer_id": "ctm_01",
			},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.MISSING_PADDLE_SUBSCRIPTION_ID_CODE, e.ErrorCode)
	}
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_SubscriptionUpdate_MissingCustomerID() {
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "subscription.updated",
		Data: map[string]any{
			"subscription": map[string]any{
				"id": "sub_01",
			},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.MISSING_PADDLE_CUSTOMER_ID_CODE, e.ErrorCode)
	}
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_SubscriptionUpdate_Success() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	org.PaddleSubscriptionID = nil
	org.SubscriptionType = constants.SUBSCRIPTION_TYPE_FREE

	s.orgRepo.On("GetByPaddleCustomerID", "ctm_01").Return(org, nil)
	s.orgRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Organization")).Run(func(args mock.Arguments) {
		updated := args.Get(1).(*models.Organization)
		s.Equal("sub_01", *updated.PaddleSubscriptionID)
		// subscription.updated does not apply plan (only subscription.created/trialing do); type stays unchanged
		s.Equal(constants.SUBSCRIPTION_TYPE_FREE, updated.SubscriptionType)
		s.NotNil(updated.SubscriptionExpiresAt)
	}).Return(nil)

	endsAt := "2025-07-15T12:00:00Z"
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "subscription.updated",
		Data: map[string]any{
			"subscription": map[string]any{
				"id":                     "sub_01",
				"customer_id":            "ctm_01",
				"items":                  []any{map[string]any{"price": map[string]any{"id": testBasicPriceID}}},
				"current_billing_period": map[string]any{"ends_at": endsAt},
			},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_SubscriptionUpdate_MultiItemOurPriceNotFirst() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	org.PaddleSubscriptionID = nil
	org.SubscriptionType = constants.SUBSCRIPTION_TYPE_FREE

	s.orgRepo.On("GetByPaddleCustomerID", "ctm_01").Return(org, nil)
	s.orgRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Organization")).Run(func(args mock.Arguments) {
		updated := args.Get(1).(*models.Organization)
		s.Equal("sub_02", *updated.PaddleSubscriptionID)
		s.Equal(constants.SUBSCRIPTION_TYPE_FREE, updated.SubscriptionType)
		s.NotNil(updated.SubscriptionExpiresAt)
	}).Return(nil)

	endsAt := "2025-07-15T12:00:00Z"
	items := []any{
		map[string]any{"price": map[string]any{"id": "pri_other_product"}},
		map[string]any{"price": map[string]any{"id": testBasicPriceID}},
	}
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "subscription.updated",
		Data: map[string]any{
			"subscription": map[string]any{
				"id":                     "sub_02",
				"customer_id":            "ctm_01",
				"items":                  items,
				"current_billing_period": map[string]any{"ends_at": endsAt},
			},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_SubscriptionUpdate_DifferentSubscriptionIDIgnored() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	storedSub := "sub_ours"
	org.PaddleSubscriptionID = &storedSub
	org.SubscriptionType = constants.SUBSCRIPTION_TYPE_BASIC

	s.orgRepo.On("GetByPaddleCustomerID", "ctm_01").Return(org, nil)

	endsAt := "2025-07-15T12:00:00Z"
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "subscription.updated",
		Data: map[string]any{
			"subscription": map[string]any{
				"id":                     "sub_other_app",
				"customer_id":            "ctm_01",
				"items":                  []any{map[string]any{"price": map[string]any{"id": testBasicPriceID}}},
				"current_billing_period": map[string]any{"ends_at": endsAt},
			},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.NoError(err)
	s.orgRepo.AssertNotCalled(s.T(), "Update", mock.Anything, mock.AnythingOfType("*models.Organization"))
	s.orgRepo.AssertExpectations(s.T())
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_TransactionCompleted_MultiItemOurPriceNotFirst() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	org.PaddleSubscriptionID = nil
	org.SubscriptionType = constants.SUBSCRIPTION_TYPE_FREE

	s.orgRepo.On("GetByPaddleCustomerID", "ctm_01").Return(org, nil)
	s.orgRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Organization")).Run(func(args mock.Arguments) {
		updated := args.Get(1).(*models.Organization)
		s.Equal(constants.SUBSCRIPTION_TYPE_BASIC, updated.SubscriptionType)
	}).Return(nil)

	endsAt := "2025-08-01T12:00:00Z"
	items := []any{
		map[string]any{"price": map[string]any{"id": "pri_addon"}},
		map[string]any{"price": map[string]any{"id": testBasicPriceID}},
	}
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "transaction.completed",
		Data: map[string]any{
			"transaction": map[string]any{
				"customer_id":     "ctm_01",
				"subscription_id": "sub_01",
				"items":           items,
				"billing_period":  map[string]any{"ends_at": endsAt},
			},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_SubscriptionCanceled_Success() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	subID := "sub_01"
	org.PaddleSubscriptionID = &subID

	s.orgRepo.On("GetByPaddleCustomerID", "ctm_01").Return(org, nil)
	s.orgRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Organization")).Run(func(args mock.Arguments) {
		updated := args.Get(1).(*models.Organization)
		s.Nil(updated.PaddleSubscriptionID)
		s.NotNil(updated.SubscriptionExpiresAt)
	}).Return(nil)

	canceledAt := "2024-07-01T00:00:00Z"
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "subscription.canceled",
		Data: map[string]any{
			"subscription": map[string]any{
				"id":          "sub_01",
				"customer_id": "ctm_01",
				"canceled_at": canceledAt,
			},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_SubscriptionCanceled_MissingCustomerID() {
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "subscription.canceled",
		Data: map[string]any{
			"subscription": map[string]any{},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.MISSING_PADDLE_CUSTOMER_ID_CODE, e.ErrorCode)
	}
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_TransactionCompleted_Success() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	org.PaddleSubscriptionID = nil
	org.SubscriptionType = constants.SUBSCRIPTION_TYPE_FREE

	s.orgRepo.On("GetByPaddleCustomerID", "ctm_01").Return(org, nil)
	s.orgRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Organization")).Run(func(args mock.Arguments) {
		updated := args.Get(1).(*models.Organization)
		s.Equal("sub_01", *updated.PaddleSubscriptionID)
		s.Equal(constants.SUBSCRIPTION_TYPE_BASIC, updated.SubscriptionType)
		s.Equal(constants.SUBSCRIPTION_TYPE_SEATS_INCLUDED_BASIC, updated.SubscriptionSeats)
		s.Equal(constants.PAYMENT_INTERVAL_MONTHLY, updated.SubscriptionPaymentInterval)
		s.NotNil(updated.SubscriptionExpiresAt)
	}).Return(nil)

	endsAt := "2025-08-01T12:00:00Z"
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "transaction.completed",
		Data: map[string]any{
			"transaction": map[string]any{
				"customer_id":     "ctm_01",
				"subscription_id": "sub_01",
				"items":           []any{map[string]any{"price": map[string]any{"id": testBasicPriceID}}},
				"billing_period":  map[string]any{"ends_at": endsAt},
			},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *PaddleServiceTestSuite) TestHandleWebhook_TransactionCompleted_MissingCustomerID() {
	payload, _ := json.Marshal(models.PaddleWebhookEvent{
		EventType: "transaction.completed",
		Data: map[string]any{
			"transaction": map[string]any{
				"customer_id": "",
			},
		},
	})
	err := s.svc.HandleWebhook(payload, "sig")
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.MISSING_PADDLE_CUSTOMER_ID_CODE, e.ErrorCode)
	}
}

// --- UpdateCustomer ---

func (s *PaddleServiceTestSuite) TestUpdateCustomer_Success() {
	org := s.orgWithPaddle(uuid.New().String(), "ctm_01")
	returned := &paddle.Customer{ID: "ctm_01", Email: org.BillingEmail}
	s.paddleClient.On("UpdateCustomer", mock.Anything, mock.AnythingOfType("*paddle.UpdateCustomerRequest")).Return(returned, nil)

	customer, err := s.svc.UpdateCustomer(org)
	s.NoError(err)
	s.Equal(returned, customer)
	s.paddleClient.AssertExpectations(s.T())
}

func (s *PaddleServiceTestSuite) TestUpdateCustomer_ClientError() {
	org := s.orgWithPaddle(uuid.New().String(), "ctm_01")
	s.paddleClient.On("UpdateCustomer", mock.Anything, mock.AnythingOfType("*paddle.UpdateCustomerRequest")).Return(&paddle.Customer{}, errors.New("api error"))

	customer, err := s.svc.UpdateCustomer(org)
	s.Error(err)
	s.Nil(customer) // service returns nil customer on error
}

// --- CreateCustomer ---

func (s *PaddleServiceTestSuite) TestCreateCustomer_Success() {
	org := s.orgWithPaddle(uuid.New().String(), "")
	org.ID = uuid.New()
	returned := &paddle.Customer{ID: "ctm_new", Email: org.BillingEmail}
	s.paddleClient.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*paddle.CreateCustomerRequest")).Return(returned, nil)

	customer, err := s.svc.CreateCustomer(org)
	s.NoError(err)
	s.Equal(returned, customer)
}

func (s *PaddleServiceTestSuite) TestCreateCustomer_GetOrCreate_ExistingCustomer() {
	org := s.orgWithPaddle(uuid.New().String(), "")
	org.ID = uuid.New()
	existingCustomer := &paddle.Customer{ID: "ctm_existing", Email: org.BillingEmail, Status: paddle.StatusActive}
	s.paddleClient.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*paddle.CreateCustomerRequest")).
		Return((*paddle.Customer)(nil), paddle.ErrCustomerAlreadyExists)
	coll := s.paddleCollectionWithCustomers(existingCustomer)
	s.paddleClient.On("ListCustomers", mock.Anything, mock.MatchedBy(func(req *paddle.ListCustomersRequest) bool {
		return len(req.Email) == 1 && req.Email[0] == org.BillingEmail
	})).Return(coll, nil)

	customer, err := s.svc.CreateCustomer(org)
	s.NoError(err)
	s.Equal(existingCustomer, customer)
}

func (s *PaddleServiceTestSuite) TestCreateCustomer_GetOrCreate_ExistingCustomer_NonSentinelError() {
	org := s.orgWithPaddle(uuid.New().String(), "")
	org.ID = uuid.New()
	existingCustomer := &paddle.Customer{ID: "ctm_existing_2", Email: org.BillingEmail, Status: paddle.StatusActive}
	s.paddleClient.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*paddle.CreateCustomerRequest")).
		Return((*paddle.Customer)(nil), errors.New("customer already exists for this email"))
	coll := s.paddleCollectionWithCustomers(existingCustomer)
	s.paddleClient.On("ListCustomers", mock.Anything, mock.MatchedBy(func(req *paddle.ListCustomersRequest) bool {
		return len(req.Email) == 1 && req.Email[0] == org.BillingEmail
	})).Return(coll, nil)

	customer, err := s.svc.CreateCustomer(org)
	s.NoError(err)
	s.Equal(existingCustomer, customer)
}

func (s *PaddleServiceTestSuite) TestCreateCustomer_GetOrCreate_ExistingCustomer_ActivatesWhenInactive() {
	org := s.orgWithPaddle(uuid.New().String(), "")
	org.ID = uuid.New()
	inactiveCustomer := &paddle.Customer{ID: "ctm_inactive", Email: org.BillingEmail, Status: paddle.StatusArchived}
	activatedCustomer := &paddle.Customer{ID: "ctm_inactive", Email: org.BillingEmail, Status: paddle.StatusActive}

	s.paddleClient.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*paddle.CreateCustomerRequest")).
		Return((*paddle.Customer)(nil), paddle.ErrCustomerAlreadyExists)
	coll := s.paddleCollectionWithCustomers(inactiveCustomer)
	s.paddleClient.On("ListCustomers", mock.Anything, mock.MatchedBy(func(req *paddle.ListCustomersRequest) bool {
		return len(req.Email) == 1 && req.Email[0] == org.BillingEmail
	})).Return(coll, nil)
	s.paddleClient.On("UpdateCustomer", mock.Anything, mock.MatchedBy(func(req *paddle.UpdateCustomerRequest) bool {
		status := req.Status.Value()
		return req.CustomerID == inactiveCustomer.ID &&
			req.Status != nil &&
			status != nil &&
			*status == paddle.StatusActive
	})).Return(activatedCustomer, nil)

	customer, err := s.svc.CreateCustomer(org)
	s.NoError(err)
	s.Equal(activatedCustomer, customer)
}

func (s *PaddleServiceTestSuite) TestCreateCustomer_GetOrCreate_ListEmpty_FetchByIDFromErrorDetail() {
	org := s.orgWithPaddle(uuid.New().String(), "")
	org.ID = uuid.New()
	apiErr := &paddleerr.Error{
		Type:   paddleerr.ErrorTypeRequestError,
		Code:   "customer_already_exists",
		Detail: "customer email conflicts with customer of id ctm_01kjcewx3d04zjbsfa8j6e0kxn",
	}
	s.paddleClient.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*paddle.CreateCustomerRequest")).
		Return((*paddle.Customer)(nil), apiErr)
	s.paddleClient.On("ListCustomers", mock.Anything, mock.Anything).Return(s.paddleCollectionWithCustomers(), nil)
	fetched := &paddle.Customer{ID: "ctm_01kjcewx3d04zjbsfa8j6e0kxn", Email: org.BillingEmail, Status: paddle.StatusActive}
	s.paddleClient.On("GetCustomer", mock.Anything, &paddle.GetCustomerRequest{CustomerID: "ctm_01kjcewx3d04zjbsfa8j6e0kxn"}).
		Return(fetched, nil)

	customer, err := s.svc.CreateCustomer(org)
	s.NoError(err)
	s.Equal(fetched, customer)
}

func (s *PaddleServiceTestSuite) TestCreateCustomer_GetOrCreate_ListFails() {
	org := s.orgWithPaddle(uuid.New().String(), "")
	org.ID = uuid.New()
	s.paddleClient.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*paddle.CreateCustomerRequest")).
		Return((*paddle.Customer)(nil), paddle.ErrCustomerAlreadyExists)
	s.paddleClient.On("ListCustomers", mock.Anything, mock.Anything).Return((*paddle.Collection[*paddle.Customer])(nil), errors.New("list error"))

	customer, err := s.svc.CreateCustomer(org)
	s.Error(err)
	s.Nil(customer)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.PADDLE_CUSTOMER_ALREADY_EXISTS_CODE, e.ErrorCode)
	}
}

// paddleCollectionWithCustomers builds a Collection from JSON so Iter returns the given customers.
func (s *PaddleServiceTestSuite) paddleCollectionWithCustomers(customers ...*paddle.Customer) *paddle.Collection[*paddle.Customer] {
	data := make([]map[string]any, 0, len(customers))
	for _, c := range customers {
		data = append(data, map[string]any{
			"id":     c.ID,
			"email":  c.Email,
			"status": string(c.Status),
		})
	}
	raw, err := json.Marshal(map[string]any{"data": data, "meta": map[string]any{"request_id": "test", "pagination": map[string]any{"per_page": 50, "next": nil, "has_more": false}}})
	s.Require().NoError(err)
	var coll paddle.Collection[*paddle.Customer]
	s.Require().NoError(coll.UnmarshalJSON(raw))
	return &coll
}

// --- CancelSubscription ---

func (s *PaddleServiceTestSuite) TestCancelSubscription_Success() {
	s.paddleClient.On("CancelSubscription", mock.Anything, mock.MatchedBy(func(req *paddle.CancelSubscriptionRequest) bool {
		return req.SubscriptionID == "sub_01"
	})).Return(&paddle.Subscription{}, nil)

	err := s.svc.CancelSubscription("sub_01")
	s.NoError(err)
}

func (s *PaddleServiceTestSuite) TestCancelSubscription_ClientError() {
	s.paddleClient.On("CancelSubscription", mock.Anything, mock.Anything).Return(&paddle.Subscription{}, errors.New("api error"))

	err := s.svc.CancelSubscription("sub_01")
	s.Error(err)
}

// --- CreateCheckoutSession ---

func (s *PaddleServiceTestSuite) TestCreateCheckoutSession_ValidationError() {
	req := paddle_service.CreateCheckoutSessionRequest{OrganizationID: "not-uuid", PlanID: "pri_01"}
	accessInfo := s.accessInfoAdmin(uuid.New().String())

	resp, err := s.svc.CreateCheckoutSession(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
}

func (s *PaddleServiceTestSuite) TestCreateCheckoutSession_OrgNotFound() {
	orgID := uuid.New().String()
	req := paddle_service.CreateCheckoutSessionRequest{OrganizationID: orgID, PlanID: "pri_01"}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(nil, errors.New("not found"))

	resp, err := s.svc.CreateCheckoutSession(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
}

func (s *PaddleServiceTestSuite) TestCreateCheckoutSession_NoPaddleCustomerID() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "")
	org.PaddleCustomerID = ""
	req := paddle_service.CreateCheckoutSessionRequest{OrganizationID: orgID, PlanID: "pri_01"}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(org, nil)

	resp, err := s.svc.CreateCheckoutSession(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
}

func (s *PaddleServiceTestSuite) TestCreateCheckoutSession_Success() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	req := paddle_service.CreateCheckoutSessionRequest{OrganizationID: orgID, PlanID: "pri_01"}
	accessInfo := s.accessInfoAdmin(orgID)
	transactionID := "txn_01abc123"

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.paddleClient.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*paddle.CreateTransactionRequest")).Return(&paddle.Transaction{
		ID: transactionID,
	}, nil)

	resp, err := s.svc.CreateCheckoutSession(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(transactionID, resp.TransactionID)
}

func (s *PaddleServiceTestSuite) TestCreateCheckoutSession_NoTransactionID() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	req := paddle_service.CreateCheckoutSessionRequest{OrganizationID: orgID, PlanID: "pri_01"}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.paddleClient.On("CreateTransaction", mock.Anything, mock.Anything).Return(&paddle.Transaction{ID: ""}, nil)

	resp, err := s.svc.CreateCheckoutSession(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
}

// --- CreateCustomerPortalSession ---

func (s *PaddleServiceTestSuite) TestCreateCustomerPortalSession_ValidationError() {
	req := paddle_service.CreateCustomerPortalSessionRequest{OrganizationID: "not-uuid"}
	accessInfo := s.accessInfoAdmin(uuid.New().String())

	resp, err := s.svc.CreateCustomerPortalSession(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
}

func (s *PaddleServiceTestSuite) TestCreateCustomerPortalSession_NoPaddleCustomerID() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "")
	org.PaddleCustomerID = ""
	req := paddle_service.CreateCustomerPortalSessionRequest{OrganizationID: orgID}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(org, nil)

	resp, err := s.svc.CreateCustomerPortalSession(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
}

func (s *PaddleServiceTestSuite) TestCreateCustomerPortalSession_Success() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	req := paddle_service.CreateCustomerPortalSessionRequest{OrganizationID: orgID}
	accessInfo := s.accessInfoAdmin(orgID)
	portalURL := "https://portal.paddle.com/overview"

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.paddleClient.On("CreateCustomerPortalSession", mock.Anything, mock.AnythingOfType("*paddle.CreateCustomerPortalSessionRequest")).Return(&paddle.CustomerPortalSession{
		URLs: paddle.CustomerPortalSessionURLs{General: paddle.CustomerPortalSessionGeneralURLs{Overview: portalURL}},
	}, nil)

	resp, err := s.svc.CreateCustomerPortalSession(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(portalURL, resp.PortalURL)
}

func (s *PaddleServiceTestSuite) TestCreateCustomerPortalSession_EmptyPortalURL() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_01")
	req := paddle_service.CreateCustomerPortalSessionRequest{OrganizationID: orgID}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.paddleClient.On("CreateCustomerPortalSession", mock.Anything, mock.Anything).Return(&paddle.CustomerPortalSession{
		URLs: paddle.CustomerPortalSessionURLs{General: paddle.CustomerPortalSessionGeneralURLs{Overview: ""}},
	}, nil)

	resp, err := s.svc.CreateCustomerPortalSession(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
}

// --- ProcessTransaction ---

func (s *PaddleServiceTestSuite) TestProcessTransaction_ValidationError() {
	orgID := uuid.New().String()
	req := &paddle_service.ProcessTransactionRequest{
		OrganizationID: "not-uuid",
		TransactionID:  "txn_01",
		PriceID:        "pri_01",
	}
	accessInfo := s.accessInfoAdmin(orgID)

	org, err := s.svc.ProcessTransaction(context.Background(), req, accessInfo)
	s.Error(err)
	s.Nil(org)
}

func (s *PaddleServiceTestSuite) TestProcessTransaction_OrgNotFound() {
	orgID := uuid.New().String()
	req := &paddle_service.ProcessTransactionRequest{
		OrganizationID: orgID,
		TransactionID:  "txn_01",
		PriceID:        testBasicPriceID,
	}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(nil, errors.New("not found"))

	org, err := s.svc.ProcessTransaction(context.Background(), req, accessInfo)
	s.Error(err)
	s.Nil(org)
}

func (s *PaddleServiceTestSuite) TestProcessTransaction_NoPaddleCustomerID() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "")
	org.PaddleCustomerID = ""
	req := &paddle_service.ProcessTransactionRequest{
		OrganizationID: orgID,
		TransactionID:  "txn_01",
		PriceID:        testBasicPriceID,
	}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(org, nil)

	orgResp, err := s.svc.ProcessTransaction(context.Background(), req, accessInfo)
	s.Error(err)
	s.Nil(orgResp)
}

func (s *PaddleServiceTestSuite) TestProcessTransaction_TransactionBelongsToDifferentCustomer() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_org")
	req := &paddle_service.ProcessTransactionRequest{
		OrganizationID: orgID,
		TransactionID:  "txn_01",
		PriceID:        testBasicPriceID,
	}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	otherCustomerID := "ctm_other"
	s.paddleClient.On("GetTransaction", mock.Anything, &paddle.GetTransactionRequest{
		TransactionID: "txn_01",
	}).Return(&paddle.Transaction{
		ID:         "txn_01",
		Status:     paddle.TransactionStatusCompleted,
		CustomerID: &otherCustomerID,
	}, nil)

	orgResp, err := s.svc.ProcessTransaction(context.Background(), req, accessInfo)
	if err != nil {
		if e, ok := err.(*yca_error.Error); ok {
			s.T().Logf("ProcessTransaction_TransactionBelongsToDifferentCustomer error: code=%s status=%d", e.ErrorCode, e.StatusCode)
		} else {
			s.T().Logf("ProcessTransaction_TransactionBelongsToDifferentCustomer error: %v", err)
		}
	}
	s.Error(err)
	s.Nil(orgResp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *PaddleServiceTestSuite) TestProcessTransaction_NotOurProduct() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_org")
	req := &paddle_service.ProcessTransactionRequest{
		OrganizationID: orgID,
		TransactionID:  "txn_01",
		PriceID:        testBasicPriceID,
	}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	customerID := "ctm_org"
	// Transaction with a different product's price (not our Basic/Pro price IDs)
	s.paddleClient.On("GetTransaction", mock.Anything, &paddle.GetTransactionRequest{
		TransactionID: "txn_01",
	}).Return(&paddle.Transaction{
		ID:         "txn_01",
		Status:     paddle.TransactionStatusCompleted,
		CustomerID: &customerID,
		Items:      []paddle.TransactionItem{{Price: paddle.Price{ID: "pri_other_product"}}},
	}, nil)

	orgResp, err := s.svc.ProcessTransaction(context.Background(), req, accessInfo)
	s.Error(err)
	s.Nil(orgResp)
}

func (s *PaddleServiceTestSuite) TestProcessTransaction_NotCompleted() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_org")
	req := &paddle_service.ProcessTransactionRequest{
		OrganizationID: orgID,
		TransactionID:  "txn_01",
		PriceID:        testBasicPriceID,
	}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	customerID := "ctm_org"
	s.paddleClient.On("GetTransaction", mock.Anything, &paddle.GetTransactionRequest{
		TransactionID: "txn_01",
	}).Return(&paddle.Transaction{
		ID:         "txn_01",
		Status:     paddle.TransactionStatus("pending"),
		CustomerID: &customerID,
	}, nil)

	orgResp, err := s.svc.ProcessTransaction(context.Background(), req, accessInfo)
	s.Error(err)
	s.Nil(orgResp)
}

func (s *PaddleServiceTestSuite) TestProcessTransaction_Success() {
	orgID := uuid.New().String()
	org := s.orgWithPaddle(orgID, "ctm_org")
	org.SubscriptionType = constants.SUBSCRIPTION_TYPE_FREE
	org.SubscriptionSeats = 1
	req := &paddle_service.ProcessTransactionRequest{
		OrganizationID: orgID,
		TransactionID:  "txn_01",
		PriceID:        testBasicPriceID,
	}
	accessInfo := s.accessInfoAdmin(orgID)

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	customerID := "ctm_org"
	subscriptionID := "sub_01"
	billingEndsAt := "2025-08-01T12:00:00Z"
	s.paddleClient.On("GetTransaction", mock.Anything, &paddle.GetTransactionRequest{
		TransactionID: "txn_01",
	}).Return(&paddle.Transaction{
		ID:             "txn_01",
		Status:         paddle.TransactionStatusCompleted,
		CustomerID:     &customerID,
		SubscriptionID: &subscriptionID,
		BillingPeriod: &paddle.TimePeriod{
			EndsAt: billingEndsAt,
		},
		Items: []paddle.TransactionItem{{Price: paddle.Price{ID: testBasicPriceID}}},
	}, nil)

	s.orgRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Organization")).Run(func(args mock.Arguments) {
		updated := args.Get(1).(*models.Organization)
		s.Equal(constants.SUBSCRIPTION_TYPE_BASIC, updated.SubscriptionType)
		s.Equal(constants.SUBSCRIPTION_TYPE_SEATS_INCLUDED_BASIC, updated.SubscriptionSeats)
		s.NotNil(updated.PaddleSubscriptionID)
		s.Equal(subscriptionID, *updated.PaddleSubscriptionID)
		s.NotNil(updated.SubscriptionExpiresAt)
	}).Return(nil)

	orgResp, err := s.svc.ProcessTransaction(context.Background(), req, accessInfo)
	if err != nil {
		if e, ok := err.(*yca_error.Error); ok {
			s.T().Logf("ProcessTransaction_Success error: code=%s status=%d", e.ErrorCode, e.StatusCode)
		} else {
			s.T().Logf("ProcessTransaction_Success error: %v", err)
		}
	}
	s.NoError(err)
	s.Require().NotNil(orgResp)
	s.Equal(constants.SUBSCRIPTION_TYPE_BASIC, orgResp.SubscriptionType)
}

// --- ApplyScheduledPlanChanges ---

func (s *PaddleServiceTestSuite) TestApplyScheduledPlanChanges_NoOrgs() {
	s.orgRepo.On("GetOrganizationsWithScheduledPlanChangeDue").Return((*[]models.Organization)(nil), nil)

	err := s.svc.ApplyScheduledPlanChanges(context.Background())
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.paddleClient.AssertNotCalled(s.T(), "UpdateSubscription")
}

func (s *PaddleServiceTestSuite) TestApplyScheduledPlanChanges_EmptyOrgs() {
	empty := []models.Organization{}
	s.orgRepo.On("GetOrganizationsWithScheduledPlanChangeDue").Return(&empty, nil)

	err := s.svc.ApplyScheduledPlanChanges(context.Background())
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.paddleClient.AssertNotCalled(s.T(), "UpdateSubscription")
}

func (s *PaddleServiceTestSuite) TestApplyScheduledPlanChanges_Success() {
	orgID := uuid.New().String()
	planID := "pri_monthly_123"
	subID := "sub_01"
	endsAt := "2025-08-01T12:00:00Z"
	org := s.orgWithPaddle(orgID, "ctm_01")
	org.PaddleSubscriptionID = &subID
	org.ScheduledPlanPriceID = &planID
	orgs := []models.Organization{*org}

	s.orgRepo.On("GetOrganizationsWithScheduledPlanChangeDue").Return(&orgs, nil)
	s.paddleClient.On("UpdateSubscription", mock.Anything, mock.MatchedBy(func(req *paddle.UpdateSubscriptionRequest) bool {
		return req != nil && req.SubscriptionID == subID
	})).Return(&paddle.Subscription{
		CurrentBillingPeriod: &paddle.TimePeriod{EndsAt: endsAt},
	}, nil)
	s.orgRepo.On("Update", nil, mock.MatchedBy(func(updated *models.Organization) bool {
		return updated != nil && updated.ID.String() == orgID && updated.ScheduledPlanPriceID == nil
	})).Return(nil)

	err := s.svc.ApplyScheduledPlanChanges(context.Background())
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.paddleClient.AssertExpectations(s.T())
}

func (s *PaddleServiceTestSuite) TestApplyScheduledPlanChanges_RepoError() {
	s.orgRepo.On("GetOrganizationsWithScheduledPlanChangeDue").Return((*[]models.Organization)(nil), errors.New("db error"))

	err := s.svc.ApplyScheduledPlanChanges(context.Background())
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *PaddleServiceTestSuite) TestApplyScheduledPlanChanges_PaddleUpdateFails_ContinuesWithoutUpdatingOrg() {
	orgID := uuid.New().String()
	planID := "pri_monthly_123"
	subID := "sub_01"
	org := s.orgWithPaddle(orgID, "ctm_01")
	org.PaddleSubscriptionID = &subID
	org.ScheduledPlanPriceID = &planID
	orgs := []models.Organization{*org}

	s.orgRepo.On("GetOrganizationsWithScheduledPlanChangeDue").Return(&orgs, nil)
	s.paddleClient.On("UpdateSubscription", mock.Anything, mock.Anything).Return((*paddle.Subscription)(nil), errors.New("paddle error"))

	err := s.svc.ApplyScheduledPlanChanges(context.Background())
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.orgRepo.AssertNotCalled(s.T(), "Update", nil, mock.Anything)
}
