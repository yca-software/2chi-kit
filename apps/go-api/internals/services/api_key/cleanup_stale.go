package api_key_service

func (s *service) CleanupStaleExpired() error {
	return s.repos.ApiKey.CleanupStaleExpired()
}
