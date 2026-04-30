package user_service

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

func (m *MockService) ChangePassword(req *ChangePasswordRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockService) Count(accessInfo *models.AccessInfo) (int, error) {
	args := m.Called(accessInfo)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockService) Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}

func (m *MockService) List(req *ListRequest, accessInfo *models.AccessInfo) (*PaginatedListResponse, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*PaginatedListResponse), args.Error(1)
}

func (m *MockService) UpdateProfile(req *UpdateProfileRequest, accessInfo *models.AccessInfo) (*models.User, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockService) UpdateLanguage(req *UpdateLanguageRequest, accessInfo *models.AccessInfo) (*models.User, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockService) AcceptTerms(req *AcceptTermsRequest, accessInfo *models.AccessInfo) (*models.User, error) {
	args := m.Called(req, accessInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockService) Get(req *GetRequest, accessInfo *models.AccessInfo) (*GetResponse, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*GetResponse), args.Error(1)
}
