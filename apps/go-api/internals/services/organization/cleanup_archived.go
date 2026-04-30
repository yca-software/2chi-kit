package organization_service

func (s *service) CleanupArchived() error {
	return s.repos.Organization.CleanupArchived()
}
