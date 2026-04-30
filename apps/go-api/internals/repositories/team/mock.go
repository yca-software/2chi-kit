package team_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.Team]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(tx yca_repository.Tx, team *models.Team) error {
	args := m.Called(tx, team)
	return args.Error(0)
}

func (m *MockRepository) GetByID(organizationID string, id string) (*models.Team, error) {
	args := m.Called(organizationID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockRepository) ListByOrganizationID(organizationID string) (*[]models.Team, error) {
	args := m.Called(organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Team), args.Error(1)
}

func (m *MockRepository) Update(tx yca_repository.Tx, team *models.Team) error {
	args := m.Called(tx, team)
	return args.Error(0)
}

func (m *MockRepository) Delete(tx yca_repository.Tx, organizationID string, id string) error {
	args := m.Called(tx, organizationID, id)
	return args.Error(0)
}
