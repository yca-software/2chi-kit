package auth_service

import (
	"strings"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_password "github.com/yca-software/go-common/password"
)

type AuthenticateWithPasswordRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"ipAddress" validate:"required,ip"`
	UserAgent string `json:"userAgent" validate:"required"`
}

func (s *service) AuthenticateWithPassword(req *AuthenticateWithPasswordRequest) (*AuthenticateResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	emailLower := strings.ToLower(req.Email)

	user, err := s.repos.User.GetByEmail(nil, emailLower)
	if err != nil {
		if e, ok := yca_error.AsError(err); ok && e.ErrorCode == constants.NOT_FOUND_CODE {
			return nil, yca_error.NewNotFoundError(nil, constants.PASSWORD_MISMATCH_CODE, nil)
		}
		return nil, err
	}

	if user.Password == nil {
		return nil, yca_error.NewNotFoundError(nil, constants.PASSWORD_MISMATCH_CODE, nil)
	}

	if isMatch := yca_password.Compare(req.Password, *user.Password); !isMatch {
		return nil, yca_error.NewNotFoundError(nil, constants.PASSWORD_MISMATCH_CODE, nil)
	}

	now := s.now()

	accessToken, err := s.generateAccessToken(user, "", "")
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
		ID:        refreshTokenID,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(s.refreshTTL) * time.Hour),
		IP:        req.IPAddress,
		UserAgent: req.UserAgent,
		TokenHash: s.hashToken(refreshToken),
	}); err != nil {
		return nil, err
	}

	return &AuthenticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
