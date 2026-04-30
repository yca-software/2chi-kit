package audit_log_service

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

func (m *MockService) Create(req *CreateRequest, accessInfo *models.AccessInfo) (*models.AuditLog, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*models.AuditLog), args.Error(1)
}

func (m *MockService) ListForOrganization(req *ListForOrganizationRequest, accessInfo *models.AccessInfo) (*ListForOrganizationResponse, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*ListForOrganizationResponse), args.Error(1)
}
