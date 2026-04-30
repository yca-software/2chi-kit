package invitation_service

func (s *service) CleanupStale() error {
	return s.repos.Invitation.CleanupStale()
}
