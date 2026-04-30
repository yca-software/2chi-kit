package auth_service

func (s *service) CleanupStalePasswordResetTokens() error {
	return s.repos.UserPasswordResetToken.Cleanup(nil)
}

func (s *service) CleanupStaleEmailVerificationTokens() error {
	return s.repos.UserEmailVerificationToken.Cleanup(nil)
}
