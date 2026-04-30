package admin_access_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.AdminAccess]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) GetByUserID(userID string) (*models.AdminAccess, error) {
	args := m.Called(userID)
	return args.Get(0).(*models.AdminAccess), args.Error(1)
}
