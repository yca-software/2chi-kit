package models

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type LocationData struct {
	Address  string
	City     string
	Zip      string
	Country  string
	PlaceID  string
	Geo      Point
	Timezone string
}

type PlaceDetails struct {
	PlaceID           string             `json:"placeId"`
	FormattedAddress  string             `json:"formattedAddress"`
	AddressComponents []AddressComponent `json:"addressComponents"`
	Geometry          PlaceGeometry      `json:"geometry"`
}

type AddressComponent struct {
	LongName  string   `json:"longName"`
	ShortName string   `json:"shortName"`
	Types     []string `json:"types"`
}

type PlaceGeometry struct {
	Location struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location"`
}

type PlacePrediction struct {
	PlaceID              string `json:"placeId"`
	Description          string `json:"description"`
	StructuredFormatting struct {
		MainText      string `json:"mainText"`
		SecondaryText string `json:"secondaryText"`
	} `json:"structuredFormatting"`
}
