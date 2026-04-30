package google_service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type AutocompleteLocationResponse struct {
	Predictions []models.PlacePrediction `json:"predictions"`
}

func (s *service) AutocompleteLocation(ctx context.Context, input string) (*AutocompleteLocationResponse, error) {
	if s.mapsAPIKey == "" {
		return nil, yca_error.NewBadRequestError(nil, constants.BAD_REQUEST_CODE, nil)
	}

	if input == "" {
		return &AutocompleteLocationResponse{Predictions: []models.PlacePrediction{}}, nil
	}

	apiURL := "https://maps.googleapis.com/maps/api/place/autocomplete/json"
	params := url.Values{}
	params.Add("input", input)
	params.Add("key", s.mapsAPIKey)
	// Use "geocode" to allow both addresses and cities (geocode includes addresses, cities, regions, etc.)
	params.Add("types", "geocode")
	params.Add("fields", "place_id,description,structured_formatting")

	reqURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.Autocomplete",
			Error:    err,
			Message:  "Failed to create request",
		})
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.Autocomplete",
			Error:    err,
			Message:  "Failed to call Google Places API",
		})
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.Autocomplete",
			Error:    err,
			Message:  "Failed to read response body",
		})
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	var googleResponse struct {
		Status      string `json:"status"`
		Predictions []struct {
			PlaceID              string `json:"place_id"`
			Description          string `json:"description"`
			StructuredFormatting struct {
				MainText      string `json:"main_text"`
				SecondaryText string `json:"secondary_text"`
			} `json:"structured_formatting"`
		} `json:"predictions"`
		ErrorMessage string `json:"error_message,omitempty"`
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.Autocomplete",
			Message:  fmt.Sprintf("Google Places API returned status %d: %s", resp.StatusCode, string(body)),
		})
		// Try to parse the error response
		if err := json.Unmarshal(body, &googleResponse); err == nil && googleResponse.ErrorMessage != "" {
			return nil, s.getErrorForStatus(googleResponse.Status)
		}
		return nil, yca_error.NewServiceUnavailableError(nil, constants.LOCATION_SEARCH_UNAVAILABLE_CODE, nil)
	}

	if err := json.Unmarshal(body, &googleResponse); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.Autocomplete",
			Error:    err,
			Message:  "Failed to parse Google Places API response",
		})
		return nil, yca_error.NewInternalServerError(err, constants.LOCATION_SEARCH_PROCESSING_ERROR_CODE, nil)
	}

	// Handle different status codes from Google Places API
	switch googleResponse.Status {
	case "OK":
		// Success - continue processing
	case "ZERO_RESULTS":
		// No results found - return empty results, not an error
		return &AutocompleteLocationResponse{Predictions: []models.PlacePrediction{}}, nil
	case "INVALID_REQUEST":
		return nil, yca_error.NewBadRequestError(nil, constants.LOCATION_SEARCH_INVALID_QUERY_CODE, nil)
	case "OVER_QUERY_LIMIT":
		return nil, yca_error.NewBadRequestError(nil, constants.LOCATION_SEARCH_TEMPORARY_ERROR_CODE, nil)
	case "REQUEST_DENIED":
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.Autocomplete",
			Message:  fmt.Sprintf("Google Places API request denied: %s", googleResponse.ErrorMessage),
		})
		return nil, yca_error.NewBadRequestError(nil, constants.LOCATION_SEARCH_DENIED_CODE, nil)
	case "UNKNOWN_ERROR":
		return nil, yca_error.NewInternalServerError(nil, constants.INTERNAL_SERVER_ERROR_CODE, nil)
	default:
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.Autocomplete",
			Message:  fmt.Sprintf("Google Places API error: %s - %s", googleResponse.Status, googleResponse.ErrorMessage),
		})
		return nil, s.getErrorForStatus(googleResponse.Status)
	}

	predictions := make([]models.PlacePrediction, len(googleResponse.Predictions))
	for i, p := range googleResponse.Predictions {
		predictions[i] = models.PlacePrediction{
			PlaceID:     p.PlaceID,
			Description: p.Description,
			StructuredFormatting: struct {
				MainText      string `json:"mainText"`
				SecondaryText string `json:"secondaryText"`
			}{
				MainText:      p.StructuredFormatting.MainText,
				SecondaryText: p.StructuredFormatting.SecondaryText,
			},
		}
	}

	return &AutocompleteLocationResponse{Predictions: predictions}, nil
}

// getErrorForStatus returns an error with the appropriate error code based on Google Places API status
func (s *service) getErrorForStatus(status string) *yca_error.Error {
	switch status {
	case "INVALID_REQUEST":
		return yca_error.NewBadRequestError(nil, constants.LOCATION_SEARCH_INVALID_QUERY_CODE, nil)
	case "OVER_QUERY_LIMIT":
		return yca_error.NewBadRequestError(nil, constants.LOCATION_SEARCH_TEMPORARY_ERROR_CODE, nil)
	case "REQUEST_DENIED":
		return yca_error.NewBadRequestError(nil, constants.LOCATION_SEARCH_DENIED_CODE, nil)
	case "UNKNOWN_ERROR":
		return yca_error.NewInternalServerError(nil, constants.INTERNAL_SERVER_ERROR_CODE, nil)
	default:
		// Fallback to generic error
		return yca_error.NewServiceUnavailableError(nil, constants.LOCATION_SEARCH_UNAVAILABLE_CODE, nil)
	}
}
