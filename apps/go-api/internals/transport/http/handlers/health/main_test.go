package health_handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	health_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/health"
)

type HealthHandlerTestSuite struct {
	suite.Suite
	handler *health_handler.Handler
	echo    *echo.Echo
}

func TestHealthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HealthHandlerTestSuite))
}

func (s *HealthHandlerTestSuite) SetupTest() {
	s.handler = health_handler.New(nil)
	s.echo = echo.New()
	s.handler.RegisterEndpoints(s.echo)
}

func (s *HealthHandlerTestSuite) TestHealth_Returns200() {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	assert.Equal(s.T(), http.StatusOK, rec.Code)
	assert.Contains(s.T(), rec.Body.String(), "ok")
	assert.Contains(s.T(), rec.Body.String(), "status")
}

func (s *HealthHandlerTestSuite) TestReady_Returns503_WhenDBIsNil() {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	assert.Equal(s.T(), http.StatusServiceUnavailable, rec.Code)
	assert.Contains(s.T(), rec.Body.String(), "unavailable")
	assert.Contains(s.T(), rec.Body.String(), "database not configured")
}
