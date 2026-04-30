package organization_member_repository

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_repository "github.com/yca-software/go-common/repository"
)

type MockRepository struct {
	yca_repository.MockRepository[models.OrganizationMember]
}

func NewMock() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(tx yca_repository.Tx, member *models.OrganizationMember) error {
	args := m.Called(tx, member)
	return args.Error(0)
}

func (m *MockRepository) GetByID(organizationID string, id string) (*models.OrganizationMember, error) {
	args := m.Called(organizationID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrganizationMember), args.Error(1)
}

func (m *MockRepository) GetByIDWithUser(organizationID string, id string) (*models.OrganizationMemberWithUser, error) {
	args := m.Called(organizationID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrganizationMemberWithUser), args.Error(1)
}

func (m *MockRepository) GetByIDWithRole(organizationID string, id string) (*models.OrganizationMemberWithOrganizationAndRole, error) {
	args := m.Called(organizationID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrganizationMemberWithOrganizationAndRole), args.Error(1)
}

func (m *MockRepository) GetByUserIDAndOrganizationID(userID string, organizationID string) (*models.OrganizationMember, error) {
	args := m.Called(userID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrganizationMember), args.Error(1)
}

func (m *MockRepository) ListByUserID(userID string) (*[]models.OrganizationMemberWithOrganization, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.OrganizationMemberWithOrganization), args.Error(1)
}

func (m *MockRepository) ListByUserIDWithRole(userID string) (*[]models.OrganizationMemberWithOrganizationAndRole, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.OrganizationMemberWithOrganizationAndRole), args.Error(1)
}

func (m *MockRepository) ListByOrganizationID(organizationID string) (*[]models.OrganizationMemberWithUser, error) {
	args := m.Called(organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.OrganizationMemberWithUser), args.Error(1)
}

func (m *MockRepository) ListUserEmailsForRole(organizationID, roleID string) ([]string, error) {
	args := m.Called(organizationID, roleID)
	var out []string
	if v := args.Get(0); v != nil {
		out = v.([]string)
	}
	return out, args.Error(1)
}

func (m *MockRepository) Update(tx yca_repository.Tx, member *models.OrganizationMember) error {
	args := m.Called(tx, member)
	return args.Error(0)
}

func (m *MockRepository) Delete(tx yca_repository.Tx, organizationID string, id string) error {
	args := m.Called(tx, organizationID, id)
	return args.Error(0)
}
