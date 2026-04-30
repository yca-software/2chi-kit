package role_repository_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	role_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/role"
	yca_error "github.com/yca-software/go-common/error"
)

type RoleRepositoryTestSuite struct {
	suite.Suite
	repo role_repository.Repository
	db   *sqlx.DB
}

func TestRoleRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RoleRepositoryTestSuite))
}

func (s *RoleRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = role_repository.New(s.db, nil)
}

func (s *RoleRepositoryTestSuite) SetupTest() {
	// Insert test organizations
	_, err := s.db.Exec(`
		INSERT INTO organizations (id, created_at, name, address, city, zip, country, place_id, geo, timezone, billing_email, custom_subscription, subscription_expires_at, subscription_type, subscription_seats, paddle_customer_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test Org 1', '123 Main St', 'New York', '10001', 'US', 'place_123', ST_MakePoint(-74.006, 40.7128), 'America/New_York', 'billing1@example.com', false, NULL, 1, 10, 'paddle_customer_1');
	`)
	require.NoError(s.T(), err)

	// Insert test roles
	permissions1, _ := json.Marshal([]string{"read", "write"})
	permissions2, _ := json.Marshal([]string{"read"})
	_, err = s.db.Exec(`
		INSERT INTO roles (id, created_at, organization_id, name, description, permissions, locked)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', '00000001-0000-0000-0000-000000000001', 'Admin', 'Admin role', $1, false),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', '00000001-0000-0000-0000-000000000001', 'User', 'User role', $2, false),
			('00000003-0000-0000-0000-000000000003', '2024-01-15 14:00:00+00', '00000001-0000-0000-0000-000000000001', 'Locked Role', 'Locked role', $1, true);
	`, permissions1, permissions2)
	require.NoError(s.T(), err)
}

func (s *RoleRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE roles CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE organizations CASCADE")
}

func (s *RoleRepositoryTestSuite) TestCreate_Success() {
	permissions := models.RolePermissions{"read", "write"}
	role := &models.Role{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Name:           "New Role",
		Description:    "New role description",
		Permissions:    permissions,
		Locked:         false,
	}
	err := s.repo.Create(nil, role)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(role.OrganizationID.String(), role.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), role.Name, got.Name)
}

func (s *RoleRepositoryTestSuite) TestGetByID_NotFound() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *RoleRepositoryTestSuite) TestGetByID_Success() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Admin", got.Name)
}

func (s *RoleRepositoryTestSuite) TestUpdate_NotFound() {
	permissions := models.RolePermissions{"read"}
	role := &models.Role{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Name:           "Updated Role",
		Description:    "Updated description",
		Permissions:    permissions,
		Locked:         false,
	}
	err := s.repo.Update(nil, role)
	require.Error(s.T(), err)
	var ycaErr *yca_error.Error
	require.True(s.T(), errors.As(err, &ycaErr), "error should be a yca_error.Error")
	require.Equal(s.T(), http.StatusNotFound, ycaErr.StatusCode)
}

func (s *RoleRepositoryTestSuite) TestUpdate_Locked() {
	permissions := models.RolePermissions{"read"}
	role := &models.Role{
		ID:             uuid.MustParse("00000003-0000-0000-0000-000000000003"),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Name:           "Updated Locked Role",
		Description:    "Updated description",
		Permissions:    permissions,
		Locked:         true,
	}
	err := s.repo.Update(nil, role)
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *RoleRepositoryTestSuite) TestUpdate_Success() {
	permissions := models.RolePermissions{"read", "write", "delete"}
	role := &models.Role{
		ID:             uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Name:           "Updated Admin",
		Description:    "Updated admin description",
		Permissions:    permissions,
		Locked:         false,
	}
	err := s.repo.Update(nil, role)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(role.OrganizationID.String(), role.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Updated Admin", got.Name)
}

func (s *RoleRepositoryTestSuite) TestDelete_NotFound() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *RoleRepositoryTestSuite) TestDelete_Locked() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", "00000003-0000-0000-0000-000000000003")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *RoleRepositoryTestSuite) TestDelete_Success() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", "00000002-0000-0000-0000-000000000002")
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000002-0000-0000-0000-000000000002")
	require.Error(s.T(), err)
	require.Nil(s.T(), got)
}

func (s *RoleRepositoryTestSuite) TestListByOrganizationID() {
	roles, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), roles)
	require.Len(s.T(), *roles, 3)
}

func (s *RoleRepositoryTestSuite) TestCreateMany_Success() {
	orgID := uuid.MustParse("00000001-0000-0000-0000-000000000001")
	roles := []models.Role{
		{
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			OrganizationID: orgID,
			Name:           "Role 1",
			Description:    "Description for Role 1",
			Permissions:    models.RolePermissions{"read", "write"},
			Locked:         false,
		},
		{
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			OrganizationID: orgID,
			Name:           "Role 2",
			Description:    "Description for Role 2",
			Permissions:    models.RolePermissions{"read"},
			Locked:         false,
		},
		{
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			OrganizationID: orgID,
			Name:           "Role 3",
			Description:    "Description for Role 3",
			Permissions:    models.RolePermissions{"read", "write", "delete"},
			Locked:         true,
		},
	}

	err := s.repo.CreateMany(nil, &roles)
	require.NoError(s.T(), err)

	// Verify all roles were created
	for _, role := range roles {
		got, err := s.repo.GetByID(role.OrganizationID.String(), role.ID.String())
		require.NoError(s.T(), err)
		require.NotNil(s.T(), got)
		require.Equal(s.T(), role.Name, got.Name)
		require.Equal(s.T(), role.Description, got.Description)
		require.Equal(s.T(), role.Permissions, got.Permissions)
		require.Equal(s.T(), role.Locked, got.Locked)
	}

	// Verify total count increased
	allRoles, err := s.repo.ListByOrganizationID(orgID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), allRoles)
	// 3 existing roles + 3 new roles = 6 total (excluding archived)
	require.Len(s.T(), *allRoles, 6)
}

func (s *RoleRepositoryTestSuite) TestCreateMany_EmptySlice() {
	roles := []models.Role{}
	err := s.repo.CreateMany(nil, &roles)
	require.NoError(s.T(), err)

	// Verify no roles were created
	allRoles, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), allRoles)
	// Should still have the 3 existing non-archived roles
	require.Len(s.T(), *allRoles, 3)
}

func (s *RoleRepositoryTestSuite) TestCreateMany_SingleRole() {
	orgID := uuid.MustParse("00000001-0000-0000-0000-000000000001")
	roles := []models.Role{
		{
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			OrganizationID: orgID,
			Name:           "Single Role",
			Description:    "Single role description",
			Permissions:    models.RolePermissions{"read"},
			Locked:         false,
		},
	}

	err := s.repo.CreateMany(nil, &roles)
	require.NoError(s.T(), err)

	// Verify the role was created
	got, err := s.repo.GetByID(orgID.String(), roles[0].ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), roles[0].Name, got.Name)
	require.Equal(s.T(), roles[0].Description, got.Description)
	require.Equal(s.T(), roles[0].Permissions, got.Permissions)
	require.Equal(s.T(), roles[0].Locked, got.Locked)
}
