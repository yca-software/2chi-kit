package location_handler_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/metrics"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	google_service "github.com/yca-software/2chi-kit/go-api/internals/services/google"
	location_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/location"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_error "github.com/yca-software/go-common/error"
)

type LocationHandlerTestSuite struct {
	suite.Suite
	handler           *location_handler.Handler
	echo              *echo.Echo
	mockGoogleService *google_service.MockService
	services          *services.Services
	middlewares       http_middlewares.Middlewares
	accessInfo        *models.AccessInfo
}

func TestLocationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LocationHandlerTestSuite))
}

func (s *LocationHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	// Set up error handler to properly handle yca_error.Error
	s.echo.HTTPErrorHandler = func(err error, c echo.Context) {
		if httpErr, ok := err.(*yca_error.Error); ok {
			c.JSON(httpErr.StatusCode, httpErr)
			return
		}
		if he, ok := err.(*echo.HTTPError); ok {
			c.JSON(he.Code, he)
			return
		}
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	s.mockGoogleService = google_service.NewMockService()
	s.services = &services.Services{
		Google: s.mockGoogleService,
	}

	mwareCfg := &http_middlewares.MiddlewaresConfig{AccessSecret: "test-secret"}
	m, _ := metrics.New("test")
	repos := &repositories.Repositories{}
	hashToken := func(raw string) string {
		h := sha256.Sum256([]byte(raw))
		return hex.EncodeToString(h[:])
	}
	s.middlewares = http_middlewares.NewMiddlewares(mwareCfg, repos, m, hashToken)

	rateLimitConfig := &http_middlewares.RateLimitConfig{}
	s.handler = location_handler.New(s.services, s.middlewares, rateLimitConfig)
	s.handler.RegisterEndpoints(s.echo.Group("/api"))

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
}

func (s *LocationHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
}

func (s *LocationHandlerTestSuite) TestAutocomplete_Success() {
	expectedResponse := &google_service.AutocompleteLocationResponse{
		Predictions: []models.PlacePrediction{},
	}

	s.mockGoogleService.On("AutocompleteLocation", mock.Anything, "test input").
		Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/location/autocomplete?input=test+input", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.Autocomplete(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockGoogleService.AssertExpectations(s.T())
}

func (s *LocationHandlerTestSuite) TestAutocomplete_MissingInput() {
	req := httptest.NewRequest(http.MethodGet, "/api/location/autocomplete", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)

	// Call handler directly to bypass middleware authentication
	err := s.handler.Autocomplete(c)

	// Handler returns error for missing input before calling service
	s.Require().Error(err)
	// Verify it's a bad request error
	var httpErr *yca_error.Error
	s.Require().True(s.ErrorAs(err, &httpErr))
	s.Equal(http.StatusBadRequest, httpErr.StatusCode)
}
