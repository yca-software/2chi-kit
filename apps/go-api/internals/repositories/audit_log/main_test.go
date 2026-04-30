package audit_log_repository_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/audit_log"
)

type AuditLogRepositoryTestSuite struct {
	suite.Suite
	repo audit_log_repository.Repository
	db   *sqlx.DB
}

func TestAuditLogRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AuditLogRepositoryTestSuite))
}

func (s *AuditLogRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = audit_log_repository.New(s.db, nil)
}

func (s *AuditLogRepositoryTestSuite) SetupTest() {
	// Insert test organizations
	_, err := s.db.Exec(`
		INSERT INTO organizations (id, created_at, name, address, city, zip, country, place_id, geo, timezone, billing_email, custom_subscription, subscription_expires_at, subscription_type, subscription_seats, paddle_customer_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test Org 1', '123 Main St', 'New York', '10001', 'US', 'place_123', ST_MakePoint(-74.006, 40.7128), 'America/New_York', 'billing1@example.com', false, NULL, 1, 10, 'paddle_customer_1');
	`)
	require.NoError(s.T(), err)

	// Insert test audit logs
	data1, _ := json.Marshal(map[string]string{"key": "value1"})
	data2, _ := json.Marshal(map[string]string{"key": "value2"})
	_, err = s.db.Exec(`
		INSERT INTO audit_logs (id, created_at, organization_id, actor_id, actor_info, impersonated_by_id, impersonated_by_email, action, resource_type, resource_id, data)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'user@example.com', NULL, '', 'create', 'user', '00000001-0000-0000-0000-000000000001', $1),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'user@example.com', NULL, '', 'update', 'user', '00000001-0000-0000-0000-000000000001', $2),
			('00000003-0000-0000-0000-000000000003', '2024-01-15 14:00:00+00', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'user@example.com', NULL, '', 'delete', 'user', '00000001-0000-0000-0000-000000000001', $1);
	`, data1, data2)
	require.NoError(s.T(), err)
}

func (s *AuditLogRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE audit_logs CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE organizations CASCADE")
}

func (s *AuditLogRepositoryTestSuite) TestCreate_Success() {
	data, _ := json.Marshal(map[string]string{"key": "value"})
	dataRaw := json.RawMessage(data)
	log := &models.AuditLog{
		ID:                  uuid.New(),
		CreatedAt:           time.Now(),
		OrganizationID:      uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		ActorID:             uuid.New(),
		ActorInfo:           "test@example.com",
		ImpersonatedByID:    uuid.NullUUID{Valid: false},
		ImpersonatedByEmail: "",
		Action:              "test_action",
		ResourceType:        "test_resource",
		ResourceID:          uuid.New(),
		Data:                &dataRaw,
	}
	err := s.repo.Create(nil, log)
	require.NoError(s.T(), err)
}

func (s *AuditLogRepositoryTestSuite) TestListByOrganizationID_NoFilters() {
	logs, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001", nil, 10, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), logs)
	require.Len(s.T(), *logs, 3)
}

func (s *AuditLogRepositoryTestSuite) TestListByOrganizationID_WithStartDate() {
	startDate := time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC)
	filters := &audit_log_repository.AuditLogFilters{
		StartDate: &startDate,
	}
	logs, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001", filters, 10, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), logs)
	require.Len(s.T(), *logs, 2)
}

func (s *AuditLogRepositoryTestSuite) TestListByOrganizationID_WithEndDate() {
	endDate := time.Date(2024, 1, 15, 13, 30, 0, 0, time.UTC)
	filters := &audit_log_repository.AuditLogFilters{
		EndDate: &endDate,
	}
	logs, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001", filters, 10, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), logs)
	require.Len(s.T(), *logs, 2)
}

func (s *AuditLogRepositoryTestSuite) TestListByOrganizationID_WithBothFilters() {
	startDate := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 15, 13, 30, 0, 0, time.UTC)
	filters := &audit_log_repository.AuditLogFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	logs, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001", filters, 10, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), logs)
	require.Len(s.T(), *logs, 1)
}

func (s *AuditLogRepositoryTestSuite) TestListByOrganizationID_Pagination() {
	logs, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001", nil, 2, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), logs)
	require.Len(s.T(), *logs, 2)

	logs2, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001", nil, 2, 2)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), logs2)
	require.Len(s.T(), *logs2, 1)
}
