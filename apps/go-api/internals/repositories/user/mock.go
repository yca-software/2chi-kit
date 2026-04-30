package user_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.User]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(tx yca_repository.Tx, user *models.User) error {
	args := m.Called(tx, user)
	return args.Error(0)
}

func (m *MockRepository) GetByEmail(tx yca_repository.Tx, email string) (*models.User, error) {
	args := m.Called(tx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetByID(tx yca_repository.Tx, id string) (*models.User, error) {
	args := m.Called(tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetByGoogleID(tx yca_repository.Tx, googleID string) (*models.User, error) {
	args := m.Called(tx, googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) Update(tx yca_repository.Tx, user *models.User) error {
	args := m.Called(tx, user)
	return args.Error(0)
}

func (m *MockRepository) Delete(tx yca_repository.Tx, user *models.User) error {
	args := m.Called(tx, user)
	return args.Error(0)
}

func (m *MockRepository) Search(searchPhrase string, limit, offset int) (*[]models.User, error) {
	args := m.Called(searchPhrase, limit, offset)
	return args.Get(0).(*[]models.User), args.Error(1)
}

func (m *MockRepository) Count() (int, error) {
	args := m.Called()
	return args.Get(0).(int), args.Error(1)
}
