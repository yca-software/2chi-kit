package team_service

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

func (m *MockService) Create(req *CreateRequest, accessInfo *models.AccessInfo) (*models.Team, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockService) Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}

func (m *MockService) List(req *ListRequest, accessInfo *models.AccessInfo) (*[]models.Team, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*[]models.Team), args.Error(1)
}

func (m *MockService) Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.Team, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*models.Team), args.Error(1)
}
