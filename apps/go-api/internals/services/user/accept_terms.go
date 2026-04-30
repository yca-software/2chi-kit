package user_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type AcceptTermsRequest struct {
	UserID       string `json:"-" validate:"required,uuid"`
	TermsVersion string `json:"termsVersion" validate:"required"`
}

func (s *service) AcceptTerms(req *AcceptTermsRequest, accessInfo *models.AccessInfo) (*models.User, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	user, err := s.repos.User.GetByID(nil, req.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOwnResource(accessInfo, user.ID.String()); err != nil {
		return nil, err
	}

	user.TermsVersion = req.TermsVersion
	user.TermsAcceptedAt = s.now()

	if err := s.repos.User.Update(nil, user); err != nil {
		return nil, err
	}

	return user, nil
}
