package google_service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
)

type MockService struct {
	mock.Mock
}

func NewMockService() *MockService {
	return &MockService{}
}

func (m *MockService) GetUserInfo(ctx context.Context, code string) (*models.GoogleUserInfo, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(*models.GoogleUserInfo), args.Error(1)
}

func (m *MockService) GetPlaceDetails(ctx context.Context, placeID string) (*PlaceDetailsResponse, error) {
	args := m.Called(ctx, placeID)
	return args.Get(0).(*PlaceDetailsResponse), args.Error(1)
}

func (m *MockService) GetLocationData(ctx context.Context, placeID string) (*models.LocationData, error) {
	args := m.Called(ctx, placeID)
	return args.Get(0).(*models.LocationData), args.Error(1)
}

func (m *MockService) AutocompleteLocation(ctx context.Context, input string) (*AutocompleteLocationResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*AutocompleteLocationResponse), args.Error(1)
}
