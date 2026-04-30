package google_service

import (
	"context"
	"time"

	"github.com/bradfitz/latlong"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
)

func (s *service) GetLocationData(ctx context.Context, placeID string) (*models.LocationData, error) {
	placeDetails, err := s.GetPlaceDetails(ctx, placeID)
	if err != nil {
		return nil, err
	}

	place := placeDetails.Result

	// Extract address components
	var city, zip, country string
	for _, component := range place.AddressComponents {
		types := component.Types
		// City: prefer locality, fallback to administrative_area_level_1 (state/province)
		if city == "" && types != nil {
			for _, t := range types {
				if t == "locality" {
					city = component.LongName
					break
				} else if city == "" && t == "administrative_area_level_1" {
					city = component.LongName
				}
			}
		}
		// Postal code
		if zip == "" && types != nil {
			for _, t := range types {
				if t == "postal_code" {
					zip = component.LongName
					break
				}
			}
		}
		// Country
		if country == "" && types != nil {
			for _, t := range types {
				if t == "country" {
					country = component.LongName
					break
				}
			}
		}
	}

	// Detect timezone from coordinates
	lat := place.Geometry.Location.Lat
	lng := place.Geometry.Location.Lng
	timezone := detectTimezoneFromCoordinates(lat, lng)

	return &models.LocationData{
		Address:  place.FormattedAddress,
		City:     city,
		Zip:      zip,
		Country:  country,
		PlaceID:  place.PlaceID,
		Geo:      models.Point{Lat: lat, Lng: lng},
		Timezone: timezone,
	}, nil
}

// detectTimezoneFromCoordinates detects the timezone from latitude and longitude coordinates
// Falls back to UTC if detection fails
func detectTimezoneFromCoordinates(lat, lng float64) string {
	timezoneName := latlong.LookupZoneName(lat, lng)
	if timezoneName == "" {
		return "UTC"
	}

	// Validate that the timezone can be loaded
	_, err := time.LoadLocation(timezoneName)
	if err != nil {
		return "UTC"
	}

	return timezoneName
}
