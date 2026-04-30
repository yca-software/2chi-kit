package organization_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.Organization]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Archive(tx yca_repository.Tx, id string) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockRepository) CleanupArchived() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRepository) Count() (int, error) {
	args := m.Called()
	return args.Get(0).(int), args.Error(1)
}

func (m *MockRepository) Create(tx yca_repository.Tx, org *models.Organization) error {
	args := m.Called(tx, org)
	return args.Error(0)
}

func (m *MockRepository) Delete(tx yca_repository.Tx, id string) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockRepository) GetByID(id string) (*models.Organization, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockRepository) GetByIDIncludeArchived(id string) (*models.Organization, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockRepository) GetByPaddleCustomerID(paddleCustomerID string) (*models.Organization, error) {
	args := m.Called(paddleCustomerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockRepository) GetOrganizationsWithScheduledPlanChangeDue() (*[]models.Organization, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Organization), args.Error(1)
}

func (m *MockRepository) Search(searchPhrase string, limit, offset int) (*[]models.Organization, error) {
	args := m.Called(searchPhrase, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Organization), args.Error(1)
}

func (m *MockRepository) SearchArchived(searchPhrase string, limit, offset int) (*[]models.Organization, error) {
	args := m.Called(searchPhrase, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Organization), args.Error(1)
}

func (m *MockRepository) Restore(tx yca_repository.Tx, id string) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockRepository) Update(tx yca_repository.Tx, org *models.Organization) error {
	args := m.Called(tx, org)
	return args.Error(0)
}
