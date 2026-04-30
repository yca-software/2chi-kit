package user_refresh_token_service

func (s *service) CleanupStaleUnused() error {
	return s.repos.UserRefreshToken.CleanupStaleUnused(nil)
}
