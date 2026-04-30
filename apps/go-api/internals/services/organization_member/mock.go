package organization_member_service

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

func (m *MockService) ListByOrganization(req *ListByOrganizationRequest, accessInfo *models.AccessInfo) (*[]models.OrganizationMemberWithUser, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*[]models.OrganizationMemberWithUser), args.Error(1)
}

func (m *MockService) ListByUser(req *ListByUserRequest, accessInfo *models.AccessInfo) (*[]models.OrganizationMemberWithOrganization, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*[]models.OrganizationMemberWithOrganization), args.Error(1)
}

func (m *MockService) Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.OrganizationMemberWithUser, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*models.OrganizationMemberWithUser), args.Error(1)
}

func (m *MockService) Remove(req *RemoveRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}
