package http_middlewares_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/metrics"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	api_key_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/api_key"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_error "github.com/yca-software/go-common/error"
)

type MiddlewaresTestSuite struct {
	suite.Suite
	cfg          *http_middlewares.MiddlewaresConfig
	metrics      *metrics.Metrics
	apiKeyMock   *api_key_repository.MockRepository
	orgMock      *organization_repository.MockRepository
	repos        *repositories.Repositories
	middlewares  http_middlewares.Middlewares
	echo         *echo.Echo
	hashToken    func(string) string
	accessSecret string
}

func TestMiddlewaresTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewaresTestSuite))
}

func (s *MiddlewaresTestSuite) SetupSuite() {
	var err error
	s.metrics, err = metrics.New("test")
	s.Require().NoError(err)
}

func (s *MiddlewaresTestSuite) SetupTest() {
	s.accessSecret = "test-access-secret"
	s.cfg = &http_middlewares.MiddlewaresConfig{AccessSecret: s.accessSecret}

	s.apiKeyMock = api_key_repository.NewMock()
	s.orgMock = organization_repository.NewMock()
	s.repos = &repositories.Repositories{ApiKey: s.apiKeyMock, Organization: s.orgMock}

	s.hashToken = func(raw string) string {
		h := sha256.Sum256([]byte(raw))
		return hex.EncodeToString(h[:])
	}

	s.middlewares = http_middlewares.NewMiddlewares(s.cfg, s.repos, s.metrics, s.hashToken)
	s.echo = echo.New()
}

func (s *MiddlewaresTestSuite) setRequestID(c echo.Context, id string) {
	c.Response().Header().Set(echo.HeaderXRequestID, id)
}

// --- AdminRoute ---

func (s *MiddlewaresTestSuite) TestAdminRoute_Returns401WhenAccessInfoIsNil() {
	handler := s.middlewares.AdminRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	var nilAccess *models.AccessInfo
	c.Set("accessInfo", nilAccess)

	err := handler(c)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &appErr))
	s.Equal(constants.UNAUTHORIZED_CODE, appErr.ErrorCode)
	s.Equal(401, appErr.StatusCode)
}

func (s *MiddlewaresTestSuite) TestAdminRoute_Returns403WhenUserIsNotAdmin() {
	handler := s.middlewares.AdminRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.Set("accessInfo", &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: false,
		},
	})

	err := handler(c)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &appErr))
	s.Equal(constants.FORBIDDEN_CODE, appErr.ErrorCode)
	s.Equal(403, appErr.StatusCode)
}

func (s *MiddlewaresTestSuite) TestAdminRoute_Returns403WhenUserIsNil() {
	handler := s.middlewares.AdminRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.Set("accessInfo", &models.AccessInfo{User: nil})

	err := handler(c)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &appErr))
	s.Equal(constants.FORBIDDEN_CODE, appErr.ErrorCode)
	s.Equal(403, appErr.StatusCode)
}

func (s *MiddlewaresTestSuite) TestAdminRoute_CallsNextWhenUserIsAdmin() {
	handler := s.middlewares.AdminRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.Set("accessInfo", &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	})

	err := handler(c)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, rec.Code)
	s.Equal("ok", rec.Body.String())
}

// --- PrivateRoute ---

func (s *MiddlewaresTestSuite) TestPrivateRoute_Returns401WhenNoApiKeyAndNoAuthorizationHeader() {
	handler := s.middlewares.PrivateRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setRequestID(c, "req-1")

	err := handler(c)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &appErr))
	s.Equal(constants.AUTHORIZATION_HEADER_REQUIRED_CODE, appErr.ErrorCode)
	s.Equal(401, appErr.StatusCode)
}

func (s *MiddlewaresTestSuite) TestPrivateRoute_Returns401WhenApiKeyNotFound() {
	s.apiKeyMock.On("GetByHash", s.hashToken("rawkey")).Return((*models.APIKey)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	defer s.apiKeyMock.AssertExpectations(s.T())

	handler := s.middlewares.PrivateRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", constants.API_KEY_PREFIX+"rawkey")
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setRequestID(c, "req-2")

	err := handler(c)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &appErr))
	s.Equal(constants.INVALID_API_KEY_CODE, appErr.ErrorCode)
	s.Equal(401, appErr.StatusCode)
}

func (s *MiddlewaresTestSuite) TestPrivateRoute_SetsAccessInfoAndCallsNextWhenApiKeyValid() {
	keyHash := s.hashToken("validrawkey")
	orgID := uuid.New()
	apiKey := &models.APIKey{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "test-key",
		KeyHash:        keyHash,
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}
	org := &models.Organization{
		ID:               orgID,
		Name:             "Test Org",
		SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO,
	}
	s.apiKeyMock.On("GetByHash", keyHash).Return(apiKey, nil)
	s.orgMock.On("GetByID", orgID.String()).Return(org, nil)
	defer s.apiKeyMock.AssertExpectations(s.T())
	defer s.orgMock.AssertExpectations(s.T())

	handler := s.middlewares.PrivateRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", constants.API_KEY_PREFIX+"validrawkey")
	req.Header.Set("User-Agent", "test-agent")
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setRequestID(c, "req-3")

	err := handler(c)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, rec.Code)
	s.Equal("ok", rec.Body.String())

	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	s.Require().NotNil(accessInfo)
	s.Equal("req-3", accessInfo.RequestID)
	s.Equal("test-agent", accessInfo.UserAgent)
	s.Require().NotNil(accessInfo.ApiKey)
	s.Equal(apiKey.ID, accessInfo.ApiKey.ID)
}

func (s *MiddlewaresTestSuite) TestPrivateRoute_ReturnsFeatureNotIncludedWhenApiKeyOrgPlanDoesNotSupportApiAccess() {
	keyHash := s.hashToken("limitedkey")
	orgID := uuid.New()
	apiKey := &models.APIKey{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "limited-key",
		KeyHash:        keyHash,
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}
	org := &models.Organization{
		ID:               orgID,
		Name:             "Limited Org",
		SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC,
	}
	s.apiKeyMock.On("GetByHash", keyHash).Return(apiKey, nil)
	s.orgMock.On("GetByID", orgID.String()).Return(org, nil)
	defer s.apiKeyMock.AssertExpectations(s.T())
	defer s.orgMock.AssertExpectations(s.T())

	handler := s.middlewares.PrivateRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", constants.API_KEY_PREFIX+"limitedkey")
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setRequestID(c, "req-api-feature-blocked")

	err := handler(c)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &appErr))
	s.Equal(constants.FEATURE_NOT_INCLUDED_CODE, appErr.ErrorCode)
	s.Equal(403, appErr.StatusCode)
}

func (s *MiddlewaresTestSuite) TestPrivateRoute_Returns401WhenBearerTokenMalformed() {
	handler := s.middlewares.PrivateRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "NotBearer token")
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setRequestID(c, "req-4")

	err := handler(c)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &appErr))
	s.Equal(constants.AUTHORIZATION_HEADER_REQUIRED_CODE, appErr.ErrorCode)
}

func (s *MiddlewaresTestSuite) TestPrivateRoute_Returns401WhenBearerTokenInvalid() {
	handler := s.middlewares.PrivateRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setRequestID(c, "req-5")

	err := handler(c)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &appErr))
	s.Equal(constants.INVALID_TOKEN_CODE, appErr.ErrorCode)
}

func (s *MiddlewaresTestSuite) TestPrivateRoute_Returns401WhenBearerTokenExpired() {
	claims := jwt.MapClaims{
		"sub":         uuid.New().String(),
		"email":       "u@example.com",
		"exp":         time.Now().Add(-time.Hour).Unix(),
		"iat":         time.Now().Add(-2 * time.Hour).Unix(),
		"permissions": []models.JWTAccessTokenPermissionData{},
		"isAdmin":     false,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, errSign := token.SignedString([]byte(s.accessSecret))
	s.Require().NoError(errSign)

	handler := s.middlewares.PrivateRoute(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setRequestID(c, "req-6")

	err := handler(c)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &appErr))
	s.Equal(constants.EXPIRED_TOKEN_CODE, appErr.ErrorCode)
}
