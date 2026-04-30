package user_refresh_token_service

import (
	"github.com/stretchr/testify/mock"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
)

type MockService struct {
	mock.Mock
}

func NewMockService() *MockService {
	return &MockService{}
}

func (m *MockService) Revoke(req *RevokeRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}

func (m *MockService) RevokeAll(req *RevokeAllRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}

func (m *MockService) ListActive(req *ListActiveRequest, accessInfo *models.AccessInfo) (*[]models.UserRefreshToken, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*[]models.UserRefreshToken), args.Error(1)
}

func (m *MockService) CleanupStaleUnused() error {
	args := m.Called()
	return args.Error(0)
}
