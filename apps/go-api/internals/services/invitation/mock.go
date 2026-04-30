package invitation_service

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

func (m *MockService) Create(req *CreateRequest, accessInfo *models.AccessInfo) (*CreateResponse, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*CreateResponse), args.Error(1)
}

func (m *MockService) Revoke(req *RevokeRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}

func (m *MockService) List(req *ListRequest, accessInfo *models.AccessInfo) (*[]models.Invitation, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*[]models.Invitation), args.Error(1)
}

func (m *MockService) CleanupStale() error {
	args := m.Called()
	return args.Error(0)
}
