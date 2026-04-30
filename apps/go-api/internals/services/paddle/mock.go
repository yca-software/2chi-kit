package paddle_service

import (
	"context"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/stretchr/testify/mock"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
)

type MockPaddleService struct {
	mock.Mock
}

func NewMockPaddleService() *MockPaddleService {
	return &MockPaddleService{}
}

func (m *MockPaddleService) CancelSubscription(paddleSubscriptionID string) error {
	args := m.Called(paddleSubscriptionID)
	return args.Error(0)
}

func (m *MockPaddleService) CreateCheckoutSession(req CreateCheckoutSessionRequest, accessInfo *models.AccessInfo) (*CheckoutSessionResponse, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*CheckoutSessionResponse), args.Error(1)
}

func (m *MockPaddleService) CreateCustomer(organization *models.Organization) (*paddle.Customer, error) {
	args := m.Called(organization)
	return args.Get(0).(*paddle.Customer), args.Error(1)
}

func (m *MockPaddleService) CreateCustomerPortalSession(req CreateCustomerPortalSessionRequest, accessInfo *models.AccessInfo) (*CustomerPortalSessionResponse, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*CustomerPortalSessionResponse), args.Error(1)
}

func (m *MockPaddleService) HandleWebhook(payload []byte, signature string) error {
	args := m.Called(payload, signature)
	return args.Error(0)
}

func (m *MockPaddleService) UpdateCustomer(organization *models.Organization) (*paddle.Customer, error) {
	args := m.Called(organization)
	return args.Get(0).(*paddle.Customer), args.Error(1)
}

func (m *MockPaddleService) ProcessTransaction(ctx context.Context, req *ProcessTransactionRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	args := m.Called(ctx, req, accessInfo)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}

	org, _ := result.(*models.Organization)
	return org, args.Error(1)
}

func (m *MockPaddleService) ChangePlan(ctx context.Context, req *ChangePlanRequest, accessInfo *models.AccessInfo) (*ChangePlanResult, error) {
	args := m.Called(ctx, req, accessInfo)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	res, _ := result.(*ChangePlanResult)
	return res, args.Error(1)
}

func (m *MockPaddleService) ApplyScheduledPlanChanges(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockPaddleClient struct {
	mock.Mock
}

func NewMockPaddleClient() *MockPaddleClient {
	return &MockPaddleClient{}
}

func (m *MockPaddleClient) CreateTransaction(ctx context.Context, req *paddle.CreateTransactionRequest) (*paddle.Transaction, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*paddle.Transaction), args.Error(1)
}

func (m *MockPaddleClient) CreateCustomer(ctx context.Context, req *paddle.CreateCustomerRequest) (*paddle.Customer, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*paddle.Customer), args.Error(1)
}

func (m *MockPaddleClient) GetCustomer(ctx context.Context, req *paddle.GetCustomerRequest) (*paddle.Customer, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paddle.Customer), args.Error(1)
}

func (m *MockPaddleClient) ListCustomers(ctx context.Context, req *paddle.ListCustomersRequest) (*paddle.Collection[*paddle.Customer], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paddle.Collection[*paddle.Customer]), args.Error(1)
}

func (m *MockPaddleClient) UpdateCustomer(ctx context.Context, req *paddle.UpdateCustomerRequest) (*paddle.Customer, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*paddle.Customer), args.Error(1)
}

func (m *MockPaddleClient) CreateCustomerPortalSession(ctx context.Context, req *paddle.CreateCustomerPortalSessionRequest) (*paddle.CustomerPortalSession, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*paddle.CustomerPortalSession), args.Error(1)
}

func (m *MockPaddleClient) CancelSubscription(ctx context.Context, req *paddle.CancelSubscriptionRequest) (*paddle.Subscription, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*paddle.Subscription), args.Error(1)
}

func (m *MockPaddleClient) GetTransaction(ctx context.Context, req *paddle.GetTransactionRequest) (*paddle.Transaction, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*paddle.Transaction), args.Error(1)
}

func (m *MockPaddleClient) GetSubscription(ctx context.Context, req *paddle.GetSubscriptionRequest) (*paddle.Subscription, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*paddle.Subscription), args.Error(1)
}

func (m *MockPaddleClient) UpdateSubscription(ctx context.Context, req *paddle.UpdateSubscriptionRequest) (*paddle.Subscription, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*paddle.Subscription), args.Error(1)
}
