package user_email_verification_token_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.UserEmailVerificationToken]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(tx yca_repository.Tx, token *models.UserEmailVerificationToken) error {
	args := m.Called(tx, token)
	return args.Error(0)
}

func (m *MockRepository) Cleanup(tx yca_repository.Tx) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockRepository) GetByHash(tx yca_repository.Tx, tokenHash string) (*models.UserEmailVerificationToken, error) {
	args := m.Called(tx, tokenHash)
	return args.Get(0).(*models.UserEmailVerificationToken), args.Error(1)
}

func (m *MockRepository) MarkAsUsed(tx yca_repository.Tx, tokenID string) error {
	args := m.Called(tx, tokenID)
	return args.Error(0)
}
