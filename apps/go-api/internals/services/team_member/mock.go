package team_member_service

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

func (m *MockService) ListByTeam(req *ListByTeamRequest, accessInfo *models.AccessInfo) (*[]models.TeamMemberWithUser, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*[]models.TeamMemberWithUser), args.Error(1)
}

func (m *MockService) Add(req *AddRequest, accessInfo *models.AccessInfo) (*models.TeamMemberWithUser, error) {
	args := m.Called(req, accessInfo)
	return args.Get(0).(*models.TeamMemberWithUser), args.Error(1)
}

func (m *MockService) Remove(req *RemoveRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}
