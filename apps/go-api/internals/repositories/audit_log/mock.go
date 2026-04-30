package audit_log_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.AuditLog]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(tx yca_repository.Tx, log *models.AuditLog) error {
	args := m.Called(tx, log)
	return args.Error(0)
}

func (m *MockRepository) ListByOrganizationID(orgID string, filters *AuditLogFilters, limit, offset int) (*[]models.AuditLog, error) {
	args := m.Called(orgID, filters, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.AuditLog), args.Error(1)
}
