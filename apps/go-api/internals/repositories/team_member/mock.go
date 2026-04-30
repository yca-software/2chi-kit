package team_member_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.TeamMember]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(tx yca_repository.Tx, member *models.TeamMember) error {
	args := m.Called(tx, member)
	return args.Error(0)
}

func (m *MockRepository) GetByID(organizationID string, id string) (*models.TeamMember, error) {
	args := m.Called(organizationID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamMember), args.Error(1)
}

func (m *MockRepository) GetByIDWithUser(tx yca_repository.Tx, organizationID string, id string) (*models.TeamMemberWithUser, error) {
	args := m.Called(tx, organizationID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamMemberWithUser), args.Error(1)
}

func (m *MockRepository) ListByUserID(userID string) (*[]models.TeamMemberWithTeam, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.TeamMemberWithTeam), args.Error(1)
}

func (m *MockRepository) ListByTeamID(organizationID string, teamID string) (*[]models.TeamMemberWithUser, error) {
	args := m.Called(organizationID, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.TeamMemberWithUser), args.Error(1)
}

func (m *MockRepository) ListByOrganizationID(organizationID string) (*[]models.TeamMemberWithUser, error) {
	args := m.Called(organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.TeamMemberWithUser), args.Error(1)
}

func (m *MockRepository) Delete(tx yca_repository.Tx, organizationID string, id string) error {
	args := m.Called(tx, organizationID, id)
	return args.Error(0)
}
