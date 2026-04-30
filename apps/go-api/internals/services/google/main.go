package google_service

import (
	"context"
	"net/http"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_log "github.com/yca-software/go-common/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Dependencies struct {
	OAuthClientID     string
	OAuthClientSecret string
	OAuthRedirectURL  string
	MapsAPIKey        string
	Logger            yca_log.Logger
	HTTPClient        *http.Client
}

type Service interface {
	AutocompleteLocation(ctx context.Context, input string) (*AutocompleteLocationResponse, error)
	GetUserInfo(ctx context.Context, code string) (*models.GoogleUserInfo, error)
	GetPlaceDetails(ctx context.Context, placeID string) (*PlaceDetailsResponse, error)
	GetLocationData(ctx context.Context, placeID string) (*models.LocationData, error)
}

type service struct {
	oauthConfig *oauth2.Config
	logger      yca_log.Logger
	httpClient  *http.Client
	mapsAPIKey  string
}

func NewService(deps *Dependencies) Service {
	googleOAuthConfig := &oauth2.Config{
		ClientID:     deps.OAuthClientID,
		ClientSecret: deps.OAuthClientSecret,
		RedirectURL:  deps.OAuthRedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	httpClient := deps.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &service{
		oauthConfig: googleOAuthConfig,
		logger:      deps.Logger,
		httpClient:  httpClient,
		mapsAPIKey:  deps.MapsAPIKey,
	}
}
