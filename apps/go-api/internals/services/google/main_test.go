package google_service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	google_service "github.com/yca-software/2chi-kit/go-api/internals/services/google"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type GoogleServiceTestSuite struct {
	suite.Suite
	svc        google_service.Service
	logger     *yca_log.MockLogger
	mockServer *httptest.Server
}

func TestGoogleServiceTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleServiceTestSuite))
}

func (s *GoogleServiceTestSuite) SetupTest() {
	s.logger = &yca_log.MockLogger{}
	// Allow any Log calls - the service logs errors and we don't need to verify them
	s.logger.On("Log", mock.Anything).Return().Maybe()
}

func (s *GoogleServiceTestSuite) TearDownTest() {
	if s.mockServer != nil {
		s.mockServer.Close()
		s.mockServer = nil
	}
}

// createServiceWithMockServer creates a service with an HTTP client that routes
// Google API requests to the mock server
func (s *GoogleServiceTestSuite) createServiceWithMockServer(handler http.HandlerFunc) {
	s.mockServer = httptest.NewServer(handler)

	// Create a custom transport that redirects Google API calls to our test server
	transport := &mockTransport{
		baseURL: s.mockServer.URL,
		handler: http.DefaultTransport,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	s.svc = google_service.NewService(&google_service.Dependencies{
		MapsAPIKey: "test-api-key",
		Logger:     s.logger,
		HTTPClient: httpClient,
	})
}

// mockTransport intercepts requests to Google APIs and routes them to the test server
type mockTransport struct {
	baseURL string
	handler http.RoundTripper
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check if this is a Google API request
	reqURL := req.URL.String()
	if strings.Contains(reqURL, "maps.googleapis.com") || strings.Contains(reqURL, "www.googleapis.com") {
		// Rewrite the URL to point to our test server
		parsedURL, err := url.Parse(t.baseURL)
		if err != nil {
			return nil, err
		}
		// Preserve the path and query from the original request
		parsedURL.Path = req.URL.Path
		parsedURL.RawQuery = req.URL.RawQuery
		req.URL = parsedURL
		req.Host = parsedURL.Host
	}

	// Use default transport for the actual request
	if t.handler == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.handler.RoundTrip(req)
}

// ============================================================================
// AutocompleteLocation Tests
// ============================================================================

func (s *GoogleServiceTestSuite) TestAutocompleteLocation_Success() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/maps/api/place/autocomplete/json", r.URL.Path)
		s.Equal("test-api-key", r.URL.Query().Get("key"))
		s.Equal("test-input", r.URL.Query().Get("input"))

		response := map[string]interface{}{
			"status": "OK",
			"predictions": []map[string]interface{}{
				{
					"place_id":    "ChIJN1t_tDeuEmsRUsoyG83frY4",
					"description": "Sydney NSW, Australia",
					"structured_formatting": map[string]interface{}{
						"main_text":      "Sydney",
						"secondary_text": "NSW, Australia",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.AutocompleteLocation(ctx, "test-input")

	s.NoError(err)
	s.NotNil(result)
	s.Len(result.Predictions, 1)
	s.Equal("ChIJN1t_tDeuEmsRUsoyG83frY4", result.Predictions[0].PlaceID)
	s.Equal("Sydney NSW, Australia", result.Predictions[0].Description)
	s.Equal("Sydney", result.Predictions[0].StructuredFormatting.MainText)
	s.Equal("NSW, Australia", result.Predictions[0].StructuredFormatting.SecondaryText)
}

func (s *GoogleServiceTestSuite) TestAutocompleteLocation_EmptyInput() {
	s.svc = google_service.NewService(&google_service.Dependencies{
		MapsAPIKey: "test-api-key",
		Logger:     s.logger,
	})

	ctx := context.Background()
	result, err := s.svc.AutocompleteLocation(ctx, "")

	s.NoError(err)
	s.NotNil(result)
	s.Empty(result.Predictions)
}

func (s *GoogleServiceTestSuite) TestAutocompleteLocation_MissingAPIKey() {
	s.svc = google_service.NewService(&google_service.Dependencies{
		MapsAPIKey: "",
		Logger:     s.logger,
	})

	ctx := context.Background()
	result, err := s.svc.AutocompleteLocation(ctx, "test-input")

	s.Error(err)
	s.Nil(result)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.BAD_REQUEST_CODE, e.ErrorCode)
	}
}

func (s *GoogleServiceTestSuite) TestAutocompleteLocation_ZeroResults() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":      "ZERO_RESULTS",
			"predictions": []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.AutocompleteLocation(ctx, "nonexistent-place-xyz")

	s.NoError(err)
	s.NotNil(result)
	s.Empty(result.Predictions)
}

func (s *GoogleServiceTestSuite) TestAutocompleteLocation_InvalidRequest() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":        "INVALID_REQUEST",
			"error_message": "Invalid request",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.AutocompleteLocation(ctx, "test-input")

	s.Error(err)
	s.Nil(result)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.LOCATION_SEARCH_INVALID_QUERY_CODE, e.ErrorCode)
	}
}

func (s *GoogleServiceTestSuite) TestAutocompleteLocation_RequestDenied() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":        "REQUEST_DENIED",
			"error_message": "Request denied",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.AutocompleteLocation(ctx, "test-input")

	s.Error(err)
	s.Nil(result)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.LOCATION_SEARCH_DENIED_CODE, e.ErrorCode)
	}
}

func (s *GoogleServiceTestSuite) TestAutocompleteLocation_Non200Status() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.AutocompleteLocation(ctx, "test-input")

	s.Error(err)
	s.Nil(result)
}

func (s *GoogleServiceTestSuite) TestAutocompleteLocation_InvalidJSON() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.AutocompleteLocation(ctx, "test-input")

	s.Error(err)
	s.Nil(result)
}

// ============================================================================
// GetPlaceDetails Tests
// ============================================================================

func (s *GoogleServiceTestSuite) TestGetPlaceDetails_Success() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/maps/api/place/details/json", r.URL.Path)
		s.Equal("test-api-key", r.URL.Query().Get("key"))
		s.Equal("test-place-id", r.URL.Query().Get("place_id"))

		response := map[string]interface{}{
			"status": "OK",
			"result": map[string]interface{}{
				"place_id":          "ChIJN1t_tDeuEmsRUsoyG83frY4",
				"formatted_address": "Sydney NSW, Australia",
				"address_components": []map[string]interface{}{
					{
						"long_name":  "Sydney",
						"short_name": "Sydney",
						"types":      []string{"locality", "political"},
					},
					{
						"long_name":  "New South Wales",
						"short_name": "NSW",
						"types":      []string{"administrative_area_level_1", "political"},
					},
					{
						"long_name":  "Australia",
						"short_name": "AU",
						"types":      []string{"country", "political"},
					},
				},
				"geometry": map[string]interface{}{
					"location": map[string]interface{}{
						"lat": -33.8688,
						"lng": 151.2093,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.GetPlaceDetails(ctx, "test-place-id")

	s.NoError(err)
	s.NotNil(result)
	s.Equal("ChIJN1t_tDeuEmsRUsoyG83frY4", result.Result.PlaceID)
	s.Equal("Sydney NSW, Australia", result.Result.FormattedAddress)
	s.Len(result.Result.AddressComponents, 3)
	s.Equal(-33.8688, result.Result.Geometry.Location.Lat)
	s.Equal(151.2093, result.Result.Geometry.Location.Lng)
}

func (s *GoogleServiceTestSuite) TestGetPlaceDetails_EmptyPlaceID() {
	s.svc = google_service.NewService(&google_service.Dependencies{
		MapsAPIKey: "test-api-key",
		Logger:     s.logger,
	})

	ctx := context.Background()
	result, err := s.svc.GetPlaceDetails(ctx, "")

	s.Error(err)
	s.Nil(result)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INTERNAL_SERVER_ERROR_CODE, e.ErrorCode)
	}
}

func (s *GoogleServiceTestSuite) TestGetPlaceDetails_MissingAPIKey() {
	s.svc = google_service.NewService(&google_service.Dependencies{
		MapsAPIKey: "",
		Logger:     s.logger,
	})

	ctx := context.Background()
	result, err := s.svc.GetPlaceDetails(ctx, "test-place-id")

	s.Error(err)
	s.Nil(result)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INTERNAL_SERVER_ERROR_CODE, e.ErrorCode)
	}
}

func (s *GoogleServiceTestSuite) TestGetPlaceDetails_NonOKStatus() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":        "REQUEST_DENIED",
			"error_message": "Request denied",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.GetPlaceDetails(ctx, "test-place-id")

	s.Error(err)
	s.Nil(result)
}

func (s *GoogleServiceTestSuite) TestGetPlaceDetails_Non200HTTPStatus() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.GetPlaceDetails(ctx, "test-place-id")

	s.Error(err)
	s.Nil(result)
}

// ============================================================================
// GetLocationData Tests
// ============================================================================

func (s *GoogleServiceTestSuite) TestGetLocationData_Success() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/maps/api/place/details/json" {
			response := map[string]interface{}{
				"status": "OK",
				"result": map[string]interface{}{
					"place_id":          "ChIJN1t_tDeuEmsRUsoyG83frY4",
					"formatted_address": "123 Main St, Sydney NSW 2000, Australia",
					"address_components": []map[string]interface{}{
						{
							"long_name":  "123",
							"short_name": "123",
							"types":      []string{"street_number"},
						},
						{
							"long_name":  "Main St",
							"short_name": "Main St",
							"types":      []string{"route"},
						},
						{
							"long_name":  "Sydney",
							"short_name": "Sydney",
							"types":      []string{"locality", "political"},
						},
						{
							"long_name":  "2000",
							"short_name": "2000",
							"types":      []string{"postal_code"},
						},
						{
							"long_name":  "Australia",
							"short_name": "AU",
							"types":      []string{"country", "political"},
						},
					},
					"geometry": map[string]interface{}{
						"location": map[string]interface{}{
							"lat": -33.8688,
							"lng": 151.2093,
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.GetLocationData(ctx, "test-place-id")

	s.NoError(err)
	s.NotNil(result)
	s.Equal("123 Main St, Sydney NSW 2000, Australia", result.Address)
	s.Equal("Sydney", result.City)
	s.Equal("2000", result.Zip)
	s.Equal("Australia", result.Country)
	s.Equal("ChIJN1t_tDeuEmsRUsoyG83frY4", result.PlaceID)
	s.Equal(-33.8688, result.Geo.Lat)
	s.Equal(151.2093, result.Geo.Lng)
	s.NotEmpty(result.Timezone) // Should detect timezone from coordinates
}

func (s *GoogleServiceTestSuite) TestGetLocationData_StateFallback() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/maps/api/place/details/json" {
			// No locality, but has administrative_area_level_1 (state)
			response := map[string]interface{}{
				"status": "OK",
				"result": map[string]interface{}{
					"place_id":          "test-id",
					"formatted_address": "New South Wales, Australia",
					"address_components": []map[string]interface{}{
						{
							"long_name":  "New South Wales",
							"short_name": "NSW",
							"types":      []string{"administrative_area_level_1", "political"},
						},
						{
							"long_name":  "Australia",
							"short_name": "AU",
							"types":      []string{"country", "political"},
						},
					},
					"geometry": map[string]interface{}{
						"location": map[string]interface{}{
							"lat": -33.8688,
							"lng": 151.2093,
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}

	s.createServiceWithMockServer(handler)

	ctx := context.Background()
	result, err := s.svc.GetLocationData(ctx, "test-place-id")

	s.NoError(err)
	s.NotNil(result)
	// Should fallback to state when no locality
	s.Equal("New South Wales", result.City)
}

func (s *GoogleServiceTestSuite) TestGetLocationData_ErrorFromGetPlaceDetails() {
	s.svc = google_service.NewService(&google_service.Dependencies{
		MapsAPIKey: "",
		Logger:     s.logger,
	})

	ctx := context.Background()
	result, err := s.svc.GetLocationData(ctx, "test-place-id")

	s.Error(err)
	s.Nil(result)
}

// ============================================================================
// GetUserInfo Tests (OAuth flow - limited testing without real OAuth server)
// ============================================================================

func (s *GoogleServiceTestSuite) TestGetUserInfo_MissingOAuthConfig() {
	s.svc = google_service.NewService(&google_service.Dependencies{
		MapsAPIKey: "",
		Logger:     s.logger,
	})

	ctx := context.Background()
	result, err := s.svc.GetUserInfo(ctx, "invalid-code")

	// This will fail because OAuth exchange requires a real OAuth server
	// We're just testing that the service handles missing config gracefully
	s.Error(err)
	s.Nil(result)
}
