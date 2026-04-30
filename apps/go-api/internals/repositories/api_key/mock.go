package api_key_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.APIKey]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(tx yca_repository.Tx, apiKey *models.APIKey) error {
	args := m.Called(tx, apiKey)
	return args.Error(0)
}

func (m *MockRepository) GetByHash(keyHash string) (*models.APIKey, error) {
	args := m.Called(keyHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}

func (m *MockRepository) GetByID(organizationID string, id string) (*models.APIKey, error) {
	args := m.Called(organizationID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}

func (m *MockRepository) ListByOrganizationID(organizationID string) (*[]models.APIKey, error) {
	args := m.Called(organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.APIKey), args.Error(1)
}

func (m *MockRepository) Delete(tx yca_repository.Tx, organizationID string, id string) error {
	args := m.Called(tx, organizationID, id)
	return args.Error(0)
}

func (m *MockRepository) Update(tx yca_repository.Tx, apiKey *models.APIKey) error {
	args := m.Called(tx, apiKey)
	return args.Error(0)
}

func (m *MockRepository) CleanupStaleExpired() error {
	args := m.Called()
	return args.Error(0)
}
