package organization_service

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

func (m *MockService) AdminCreateOrganizationWithCustomSubscription(req *AdminCreateOrganizationWithCustomSubscriptionRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	args := m.Called(req, accessInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockService) AdminUpdateSubscriptionSettings(req *AdminUpdateSubscriptionSettingsRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	args := m.Called(req, accessInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockService) Archive(req *ArchiveRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}

func (m *MockService) CleanupArchived() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockService) Count(accessInfo *models.AccessInfo) (int, error) {
	args := m.Called(accessInfo)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockService) Create(req *CreateRequest, accessInfo *models.AccessInfo) (*CreateResponse, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*CreateResponse), args.Error(1)
}

func (m *MockService) Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}

func (m *MockService) Get(req *GetRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockService) GetArchived(req *GetRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	args := m.Called(req, accessInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockService) List(req *ListRequest, accessInfo *models.AccessInfo) (*PaginatedListResponse, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*PaginatedListResponse), args.Error(1)
}

func (m *MockService) ListArchived(req *ListRequest, accessInfo *models.AccessInfo) (*PaginatedListResponse, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*PaginatedListResponse), args.Error(1)
}

func (m *MockService) Restore(req *RestoreRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}

func (m *MockService) Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*models.Organization), args.Error(1)
}
