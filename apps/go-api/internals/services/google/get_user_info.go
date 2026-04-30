package google_service

import (
	"context"
	"encoding/json"
	"io"

	"github.com/yca-software/2chi-kit/go-api/internals/models"
)

func (s *service) GetUserInfo(ctx context.Context, code string) (*models.GoogleUserInfo, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUser models.GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	return &googleUser, nil
}
