package user_service

import "github.com/yca-software/2chi-kit/go-api/internals/models"

func (s *service) Count(accessInfo *models.AccessInfo) (int, error) {
	if err := s.authorizer.CheckAdmin(accessInfo); err != nil {
		return 0, err
	}

	return s.repos.User.Count()
}
