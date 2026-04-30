package invitation_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.Invitation]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(tx yca_repository.Tx, invitation *models.Invitation) error {
	args := m.Called(tx, invitation)
	return args.Error(0)
}

func (m *MockRepository) GetByID(organizationID string, id string) (*models.Invitation, error) {
	args := m.Called(organizationID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invitation), args.Error(1)
}

func (m *MockRepository) GetByTokenHash(tokenHash string) (*models.Invitation, error) {
	args := m.Called(tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invitation), args.Error(1)
}

func (m *MockRepository) Update(tx yca_repository.Tx, invitation *models.Invitation) error {
	args := m.Called(tx, invitation)
	return args.Error(0)
}

func (m *MockRepository) ListByOrganizationID(organizationID string) (*[]models.Invitation, error) {
	args := m.Called(organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Invitation), args.Error(1)
}

func (m *MockRepository) CleanupStale() error {
	args := m.Called()
	return args.Error(0)
}
