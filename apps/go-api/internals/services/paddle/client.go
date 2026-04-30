package paddle_service

import (
	"context"
	"fmt"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
)

type PaddleClient interface {
	CreateTransaction(ctx context.Context, req *paddle.CreateTransactionRequest) (*paddle.Transaction, error)
	CreateCustomer(ctx context.Context, req *paddle.CreateCustomerRequest) (*paddle.Customer, error)
	GetCustomer(ctx context.Context, req *paddle.GetCustomerRequest) (*paddle.Customer, error)
	ListCustomers(ctx context.Context, req *paddle.ListCustomersRequest) (*paddle.Collection[*paddle.Customer], error)
	UpdateCustomer(ctx context.Context, req *paddle.UpdateCustomerRequest) (*paddle.Customer, error)
	CreateCustomerPortalSession(ctx context.Context, req *paddle.CreateCustomerPortalSessionRequest) (*paddle.CustomerPortalSession, error)
	CancelSubscription(ctx context.Context, req *paddle.CancelSubscriptionRequest) (*paddle.Subscription, error)
	GetTransaction(ctx context.Context, req *paddle.GetTransactionRequest) (*paddle.Transaction, error)
	GetSubscription(ctx context.Context, req *paddle.GetSubscriptionRequest) (*paddle.Subscription, error)
	UpdateSubscription(ctx context.Context, req *paddle.UpdateSubscriptionRequest) (*paddle.Subscription, error)
}

type DefaultPaddleClient struct {
	SDK *paddle.SDK
}

func (c *DefaultPaddleClient) CreateTransaction(ctx context.Context, req *paddle.CreateTransactionRequest) (*paddle.Transaction, error) {
	return c.SDK.TransactionsClient.CreateTransaction(ctx, req)
}

func (c *DefaultPaddleClient) CreateCustomer(ctx context.Context, req *paddle.CreateCustomerRequest) (*paddle.Customer, error) {
	return c.SDK.CustomersClient.CreateCustomer(ctx, req)
}

func (c *DefaultPaddleClient) GetCustomer(ctx context.Context, req *paddle.GetCustomerRequest) (*paddle.Customer, error) {
	return c.SDK.CustomersClient.GetCustomer(ctx, req)
}

func (c *DefaultPaddleClient) ListCustomers(ctx context.Context, req *paddle.ListCustomersRequest) (*paddle.Collection[*paddle.Customer], error) {
	return c.SDK.CustomersClient.ListCustomers(ctx, req)
}

func (c *DefaultPaddleClient) UpdateCustomer(ctx context.Context, req *paddle.UpdateCustomerRequest) (*paddle.Customer, error) {
	return c.SDK.CustomersClient.UpdateCustomer(ctx, req)
}

func (c *DefaultPaddleClient) CreateCustomerPortalSession(ctx context.Context, req *paddle.CreateCustomerPortalSessionRequest) (*paddle.CustomerPortalSession, error) {
	return c.SDK.CustomerPortalSessionsClient.CreateCustomerPortalSession(ctx, req)
}

func (c *DefaultPaddleClient) CancelSubscription(ctx context.Context, req *paddle.CancelSubscriptionRequest) (*paddle.Subscription, error) {
	return c.SDK.SubscriptionsClient.CancelSubscription(ctx, req)
}

func (c *DefaultPaddleClient) GetTransaction(ctx context.Context, req *paddle.GetTransactionRequest) (*paddle.Transaction, error) {
	return c.SDK.TransactionsClient.GetTransaction(ctx, req)
}

func (c *DefaultPaddleClient) GetSubscription(ctx context.Context, req *paddle.GetSubscriptionRequest) (*paddle.Subscription, error) {
	return c.SDK.SubscriptionsClient.GetSubscription(ctx, req)
}

func (c *DefaultPaddleClient) UpdateSubscription(ctx context.Context, req *paddle.UpdateSubscriptionRequest) (*paddle.Subscription, error) {
	return c.SDK.SubscriptionsClient.UpdateSubscription(ctx, req)
}

func NewClient(apiKey string, environment string) (PaddleClient, error) {
	if apiKey == "" {
		return nil, nil
	}

	baseURL := paddle.SandboxBaseURL
	if environment == "production" {
		baseURL = paddle.ProductionBaseURL
	}

	sdk, err := paddle.New(apiKey, paddle.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Paddle SDK: %w", err)
	}

	return &DefaultPaddleClient{SDK: sdk}, nil
}
