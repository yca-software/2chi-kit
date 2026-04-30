package user_refresh_token_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.UserRefreshToken]
}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(tx yca_repository.Tx, token *models.UserRefreshToken) error {
	args := m.Called(tx, token)
	return args.Error(0)
}

func (m *MockRepository) CleanupStaleUnused(tx yca_repository.Tx) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockRepository) GetByHash(tx yca_repository.Tx, tokenHash string) (*models.UserRefreshToken, error) {
	args := m.Called(tx, tokenHash)
	return args.Get(0).(*models.UserRefreshToken), args.Error(1)
}

func (m *MockRepository) GetActiveByUserID(tx yca_repository.Tx, userID string) (*[]models.UserRefreshToken, error) {
	args := m.Called(tx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.UserRefreshToken), args.Error(1)
}

func (m *MockRepository) GetActiveImpersonationTokenByUserID(tx yca_repository.Tx, userID string) (*models.UserRefreshToken, error) {
	args := m.Called(tx, userID)
	return args.Get(0).(*models.UserRefreshToken), args.Error(1)
}

func (m *MockRepository) Revoke(tx yca_repository.Tx, userID string, tokenID string) error {
	args := m.Called(tx, userID, tokenID)
	return args.Error(0)
}

func (m *MockRepository) RevokeByHash(tx yca_repository.Tx, tokenHash string) error {
	args := m.Called(tx, tokenHash)
	return args.Error(0)
}

func (m *MockRepository) RevokeAll(tx yca_repository.Tx, userID string) error {
	args := m.Called(tx, userID)
	return args.Error(0)
}

func (m *MockRepository) RevokeAllExcept(tx yca_repository.Tx, userID string, excludeTokenID string) error {
	args := m.Called(tx, userID, excludeTokenID)
	return args.Error(0)
}
