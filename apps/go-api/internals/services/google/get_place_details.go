package google_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type PlaceDetailsResponse struct {
	Result models.PlaceDetails `json:"result"`
}

func (s *service) GetPlaceDetails(ctx context.Context, placeID string) (*PlaceDetailsResponse, error) {
	if s.mapsAPIKey == "" {
		return nil, yca_error.NewInternalServerError(nil, constants.INTERNAL_SERVER_ERROR_CODE, nil)
	}

	if placeID == "" {
		return nil, yca_error.NewInternalServerError(nil, constants.INTERNAL_SERVER_ERROR_CODE, nil)
	}

	apiURL := "https://maps.googleapis.com/maps/api/place/details/json"
	params := url.Values{}
	params.Add("place_id", placeID)
	params.Add("key", s.mapsAPIKey)
	params.Add("fields", "place_id,formatted_address,address_components,geometry")

	reqURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.GetPlaceDetails",
			Error:    err,
			Message:  "Failed to create request",
		})
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.GetPlaceDetails",
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
			Location: "locationService.GetPlaceDetails",
			Error:    err,
			Message:  "Failed to read response body",
		})
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.GetPlaceDetails",
			Message:  fmt.Sprintf("Google Places API returned status %d: %s", resp.StatusCode, string(body)),
		})
		return nil, yca_error.NewInternalServerError(nil, constants.INTERNAL_SERVER_ERROR_CODE, nil)
	}

	var googleResponse struct {
		Status string `json:"status"`
		Result struct {
			PlaceID           string `json:"place_id"`
			FormattedAddress  string `json:"formatted_address"`
			AddressComponents []struct {
				LongName  string   `json:"long_name"`
				ShortName string   `json:"short_name"`
				Types     []string `json:"types"`
			} `json:"address_components"`
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"result"`
		ErrorMessage string `json:"error_message,omitempty"`
	}

	if err := json.Unmarshal(body, &googleResponse); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.GetPlaceDetails",
			Error:    err,
			Message:  "Failed to parse Google Places API response",
		})
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	if googleResponse.Status != "OK" {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "locationService.GetPlaceDetails",
			Message:  fmt.Sprintf("Google Places API error: %s - %s", googleResponse.Status, googleResponse.ErrorMessage),
		})
		return nil, yca_error.NewInternalServerError(errors.New(googleResponse.ErrorMessage), constants.INTERNAL_SERVER_ERROR_CODE, nil)
	}

	addressComponents := make([]models.AddressComponent, len(googleResponse.Result.AddressComponents))
	for i, ac := range googleResponse.Result.AddressComponents {
		addressComponents[i] = models.AddressComponent{
			LongName:  ac.LongName,
			ShortName: ac.ShortName,
			Types:     ac.Types,
		}
	}

	return &PlaceDetailsResponse{
		Result: models.PlaceDetails{
			PlaceID:           googleResponse.Result.PlaceID,
			FormattedAddress:  googleResponse.Result.FormattedAddress,
			AddressComponents: addressComponents,
			Geometry: models.PlaceGeometry{
				Location: struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				}{
					Lat: googleResponse.Result.Geometry.Location.Lat,
					Lng: googleResponse.Result.Geometry.Location.Lng,
				},
			},
		},
	}, nil
}
