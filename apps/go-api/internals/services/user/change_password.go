package user_service

import (
	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	yca_error "github.com/yca-software/go-common/error"
	yca_password "github.com/yca-software/go-common/password"
)

type ChangePasswordRequest struct {
	UserID          string `json:"-" validate:"required,uuid"`
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

func (s *service) ChangePassword(req *ChangePasswordRequest) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", nil)
	}

	user, err := s.repos.User.GetByID(nil, userID.String())
	if err != nil {
		if e, ok := err.(*yca_error.Error); ok {
			if e.ErrorCode == constants.NOT_FOUND_CODE {
				return yca_error.NewUnauthorizedError(nil, "", nil)
			}
		}
		return err
	}

	if user.Password != nil {
		if isMatch := yca_password.Compare(req.CurrentPassword, *user.Password); !isMatch {
			return yca_error.NewUnprocessableEntityError(nil, constants.CURRENT_PASSWORD_MISMATCH_CODE, nil)
		}
	}

	hashedPassword, err := s.passwordHashFn(req.NewPassword)
	if err != nil {
		return err
	}
	user.Password = &hashedPassword

	if err := s.repos.User.Update(nil, user); err != nil {
		return err
	}

	return nil
}
