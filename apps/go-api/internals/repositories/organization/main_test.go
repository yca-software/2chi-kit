package organization_repository_test

import (
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
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	yca_error "github.com/yca-software/go-common/error"
)

func stringPtr(s string) *string {
	return &s
}

type OrganizationRepositoryTestSuite struct {
	suite.Suite
	repo organization_repository.Repository
	db   *sqlx.DB
}

func TestOrganizationRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationRepositoryTestSuite))
}

func (s *OrganizationRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = organization_repository.New(s.db, nil)
}

func (s *OrganizationRepositoryTestSuite) SetupTest() {
	_, err := s.db.Exec(`
		INSERT INTO organizations (id, created_at, name, address, city, zip, country, place_id, geo, timezone, billing_email, custom_subscription, subscription_expires_at, subscription_type, subscription_seats, paddle_customer_id, paddle_subscription_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test Org 1', '123 Main St', 'New York', '10001', 'US', 'place_123', ST_MakePoint(-74.006, 40.7128), 'America/New_York', 'billing1@example.com', false, NULL, 1, 10, 'paddle_customer_1', 'paddle_sub_1'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', 'Test Org 2', '456 Oak Ave', 'Los Angeles', '90001', 'US', 'place_456', ST_MakePoint(-118.2437, 34.0522), 'America/Los_Angeles', 'billing2@example.com', true, '2025-12-31 23:59:59+00', 2, 20, 'paddle_customer_2', NULL),
			('00000003-0000-0000-0000-000000000003', '2024-01-15 14:00:00+00', 'Alpha Company', '789 Pine Rd', 'Chicago', '60601', 'US', 'place_789', ST_MakePoint(-87.6298, 41.8781), 'America/Chicago', 'billing3@example.com', false, NULL, 1, 5, 'paddle_customer_3', 'paddle_sub_3'),
			('00000004-0000-0000-0000-000000000004', '2024-01-15 15:00:00+00', 'Beta Industries', '321 Elm St', 'Houston', '77001', 'US', 'place_321', ST_MakePoint(-95.3698, 29.7604), 'America/Chicago', 'billing4@example.com', false, NULL, 1, 15, 'paddle_customer_4', NULL),
			('00000005-0000-0000-0000-000000000005', '2024-01-10 10:00:00+00', 'Archived Org', '999 Old St', 'Miami', '33101', 'US', 'place_999', ST_MakePoint(-80.1918, 25.7617), 'America/New_York', 'billing5@example.com', false, NULL, 1, 1, 'paddle_customer_5', NULL);
		
		UPDATE organizations SET deleted_at = '2024-01-12 10:00:00+00' WHERE id = '00000005-0000-0000-0000-000000000005';
	`)
	require.NoError(s.T(), err)
}

func (s *OrganizationRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE organizations CASCADE")
}

func (s *OrganizationRepositoryTestSuite) TestCreate_DuplicatePaddleCustomerID() {
	paddleCustomerID := "paddle_customer_1"
	org := &models.Organization{
		ID:                 uuid.New(),
		CreatedAt:          time.Now(),
		Name:               "Duplicate Org",
		Address:            "123 Test St",
		City:               "Test City",
		Zip:                "12345",
		Country:            "US",
		PlaceID:            "place_test",
		Geo:                models.Point{},
		Timezone:           "America/New_York",
		BillingEmail:       "test@example.com",
		CustomSubscription: false,
		SubscriptionType:   1,
		SubscriptionSeats:  10,
		PaddleCustomerID:   paddleCustomerID,
	}
	err := s.repo.Create(nil, org)
	require.Error(s.T(), err)
	require.Equal(s.T(), 409, err.(*yca_error.Error).StatusCode)
}

func (s *OrganizationRepositoryTestSuite) TestCreate_Success() {
	paddleCustomerID := "paddle_customer_new"
	org := &models.Organization{
		ID:                 uuid.New(),
		CreatedAt:          time.Now(),
		Name:               "New Org",
		Address:            "123 New St",
		City:               "New City",
		Zip:                "54321",
		Country:            "US",
		PlaceID:            "place_new",
		Geo:                models.Point{},
		Timezone:           "America/New_York",
		BillingEmail:       "new@example.com",
		CustomSubscription: false,
		SubscriptionType:   1,
		SubscriptionSeats:  10,
		PaddleCustomerID:   paddleCustomerID,
	}
	err := s.repo.Create(nil, org)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(org.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), org.Name, got.Name)
}

func (s *OrganizationRepositoryTestSuite) TestGetByID_NotFound() {
	got, err := s.repo.GetByID(uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *OrganizationRepositoryTestSuite) TestGetByID_Archived() {
	got, err := s.repo.GetByID("00000005-0000-0000-0000-000000000005")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *OrganizationRepositoryTestSuite) TestGetByID_Success() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Test Org 1", got.Name)
}

func (s *OrganizationRepositoryTestSuite) TestGetByPaddleCustomerID_NotFound() {
	got, err := s.repo.GetByPaddleCustomerID("nonexistent_paddle_customer")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *OrganizationRepositoryTestSuite) TestGetByPaddleCustomerID_Archived() {
	got, err := s.repo.GetByPaddleCustomerID("paddle_customer_5")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *OrganizationRepositoryTestSuite) TestGetByPaddleCustomerID_Success() {
	got, err := s.repo.GetByPaddleCustomerID("paddle_customer_1")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Test Org 1", got.Name)
}

func (s *OrganizationRepositoryTestSuite) TestUpdate_NotFound() {
	org := &models.Organization{
		ID:                 uuid.New(),
		CreatedAt:          time.Now(),
		Name:               "Updated Org",
		Address:            "123 Updated St",
		City:               "Updated City",
		Zip:                "12345",
		Country:            "US",
		PlaceID:            "place_updated",
		Geo:                models.Point{},
		Timezone:           "America/New_York",
		BillingEmail:       "updated@example.com",
		CustomSubscription: false,
		SubscriptionType:   1,
		SubscriptionSeats:  10,
	}
	err := s.repo.Update(nil, org)
	require.Error(s.T(), err)
	var ycaErr *yca_error.Error
	require.True(s.T(), errors.As(err, &ycaErr), "error should be a yca_error.Error")
	require.Equal(s.T(), http.StatusNotFound, ycaErr.StatusCode)
}

func (s *OrganizationRepositoryTestSuite) TestUpdate_Archived() {
	org := &models.Organization{
		ID:                 uuid.MustParse("00000005-0000-0000-0000-000000000005"),
		CreatedAt:          time.Now(),
		Name:               "Updated Archived Org",
		Address:            "123 Updated St",
		City:               "Updated City",
		Zip:                "12345",
		Country:            "US",
		PlaceID:            "place_updated",
		Geo:                models.Point{},
		Timezone:           "America/New_York",
		BillingEmail:       "updated@example.com",
		CustomSubscription: false,
		SubscriptionType:   1,
		SubscriptionSeats:  10,
	}
	err := s.repo.Update(nil, org)
	require.Error(s.T(), err)
	var ycaErr *yca_error.Error
	require.True(s.T(), errors.As(err, &ycaErr), "error should be a yca_error.Error")
	require.Equal(s.T(), http.StatusNotFound, ycaErr.StatusCode)
}

func (s *OrganizationRepositoryTestSuite) TestUpdate_Success() {
	org := &models.Organization{
		ID:                   uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt:            time.Now(),
		Name:                 "Updated Test Org 1",
		Address:              "123 Updated Main St",
		City:                 "Updated New York",
		Zip:                  "10002",
		Country:              "US",
		PlaceID:              "place_updated_123",
		Geo:                  models.Point{},
		Timezone:             "America/New_York",
		BillingEmail:         "updated_billing1@example.com",
		CustomSubscription:   true,
		SubscriptionType:     2,
		SubscriptionSeats:    25,
		PaddleSubscriptionID: stringPtr("paddle_sub_updated"),
	}
	err := s.repo.Update(nil, org)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(org.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Updated Test Org 1", got.Name)
	require.Equal(s.T(), "123 Updated Main St", got.Address)
}

func (s *OrganizationRepositoryTestSuite) TestDelete_NotFound() {
	err := s.repo.Delete(nil, uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *OrganizationRepositoryTestSuite) TestDelete_Success() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001")
	require.Error(s.T(), err)
	require.Nil(s.T(), got)
}

func (s *OrganizationRepositoryTestSuite) TestArchive_NotFound() {
	err := s.repo.Archive(nil, uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *OrganizationRepositoryTestSuite) TestArchive_AlreadyArchived() {
	err := s.repo.Archive(nil, "00000005-0000-0000-0000-000000000005")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *OrganizationRepositoryTestSuite) TestArchive_Success() {
	err := s.repo.Archive(nil, "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001")
	require.Error(s.T(), err)
	require.Nil(s.T(), got)
}

func (s *OrganizationRepositoryTestSuite) TestSearch_Empty() {
	orgs, err := s.repo.Search("", 10, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), orgs)
	require.Len(s.T(), *orgs, 4) // Only non-archived organizations
}

func (s *OrganizationRepositoryTestSuite) TestSearch_WithPhrase() {
	orgs, err := s.repo.Search("Test", 10, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), orgs)
	require.Len(s.T(), *orgs, 2)
	require.Equal(s.T(), "Test Org 2", (*orgs)[0].Name)
	require.Equal(s.T(), "Test Org 1", (*orgs)[1].Name)
}

func (s *OrganizationRepositoryTestSuite) TestSearch_WithAddressPhrase() {
	orgs, err := s.repo.Search("Main", 10, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), orgs)
	require.Len(s.T(), *orgs, 1)
	require.Equal(s.T(), "Test Org 1", (*orgs)[0].Name)
}

func (s *OrganizationRepositoryTestSuite) TestSearch_Pagination() {
	orgs, err := s.repo.Search("", 2, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), orgs)
	require.Len(s.T(), *orgs, 2)

	orgs2, err := s.repo.Search("", 2, 2)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), orgs2)
	require.Len(s.T(), *orgs2, 2)
}

func (s *OrganizationRepositoryTestSuite) TestCount() {
	count, err := s.repo.Count()
	require.NoError(s.T(), err)
	require.Equal(s.T(), 4, count) // Only non-archived organizations
}

func (s *OrganizationRepositoryTestSuite) TestCleanupArchived() {
	// First, archive an organization with old deleted_at
	oldTime := time.Now().Add(-31 * 24 * time.Hour)
	_, err := s.db.Exec(`
		UPDATE organizations 
		SET deleted_at = $1 
		WHERE id = '00000002-0000-0000-0000-000000000002'
	`, oldTime)
	require.NoError(s.T(), err)

	// Run cleanup
	err = s.repo.CleanupArchived()
	require.NoError(s.T(), err)

	// Verify the organization was deleted
	var count int
	err = s.db.Get(&count, "SELECT COUNT(*) FROM organizations WHERE id = '00000002-0000-0000-0000-000000000002'")
	require.NoError(s.T(), err)
	require.Equal(s.T(), 0, count)

	// Verify the old archived org (00000005) was also deleted
	err = s.db.Get(&count, "SELECT COUNT(*) FROM organizations WHERE id = '00000005-0000-0000-0000-000000000005'")
	require.NoError(s.T(), err)
	require.Equal(s.T(), 0, count)
}
