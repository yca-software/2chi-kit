package auth_service

import (
	"time"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type ImpersonateRequest struct {
	UserID    string `json:"-" validate:"required,uuid"`
	IPAddress string `json:"ipAddress" validate:"required,ip"`
	UserAgent string `json:"userAgent" validate:"required"`
}

func (s *service) Impersonate(req *ImpersonateRequest, accessInfo *models.AccessInfo) (*AuthenticateResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	if err := s.authorizer.CheckAdmin(accessInfo); err != nil {
		return nil, err
	}

	user, err := s.repos.User.GetByID(nil, req.UserID)
	if err != nil {
		return nil, err
	}

	now := s.now()

	accessToken, err := s.generateAccessToken(user, accessInfo.User.UserID.String(), accessInfo.User.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	refreshTokenID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	if err = s.repos.UserRefreshToken.Create(nil, &models.UserRefreshToken{
		ID:             refreshTokenID,
		UserID:         user.ID,
		CreatedAt:      now,
		ExpiresAt:      now.Add(1 * time.Hour),
		IP:             req.IPAddress,
		UserAgent:      req.UserAgent,
		TokenHash:      s.hashToken(refreshToken),
		ImpersonatedBy: uuid.NullUUID{UUID: accessInfo.User.UserID, Valid: true},
	}); err != nil {
		return nil, err
	}

	return &AuthenticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
