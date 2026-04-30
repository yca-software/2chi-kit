package jobs

// TODO

/* type CleanupHandlerSuite struct {
	suite.Suite
}

func TestCleanupHandlerSuite(t *testing.T) {
	suite.Run(t, new(CleanupHandlerSuite))
}

func (s *CleanupHandlerSuite) TestAllStepsSucceed() {
	orgMock := organization_service.NewMockService()
	urtMock := user_refresh_token_service.NewMockService()
	authMock := auth_service.NewMockService()
	invMock := invitation_service.NewMockService()
	apiKeyMock := api_key_service.NewMockService()

	orgMock.On("CleanupArchived").Return(nil).Once()
	urtMock.On("CleanupStaleUnused").Return(nil).Once()
	authMock.On("CleanupStalePasswordResetTokens").Return(nil).Once()
	authMock.On("CleanupStaleEmailVerificationTokens").Return(nil).Once()
	invMock.On("CleanupStale").Return(nil).Once()
	apiKeyMock.On("CleanupStaleExpired").Return(nil).Once()

	srvs := &services.Services{
		Organization:     orgMock,
		UserRefreshToken: urtMock,
		Auth:             authMock,
		Invitation:       invMock,
		ApiKey:           apiKeyMock,
	}
	jc := &Client{log: yca_log.New()}
	c := NewConsumers(srvs, jc, yca_log.New())

	err := c.cleanupHandler()()
	s.Require().NoError(err)

	orgMock.AssertExpectations(s.T())
	urtMock.AssertExpectations(s.T())
	authMock.AssertExpectations(s.T())
	invMock.AssertExpectations(s.T())
	apiKeyMock.AssertExpectations(s.T())
}

func (s *CleanupHandlerSuite) TestStepErrorStillReturnsNil() {
	orgMock := organization_service.NewMockService()
	urtMock := user_refresh_token_service.NewMockService()
	authMock := auth_service.NewMockService()
	invMock := invitation_service.NewMockService()
	apiKeyMock := api_key_service.NewMockService()

	orgMock.On("CleanupArchived").Return(errors.New("boom")).Once()
	urtMock.On("CleanupStaleUnused").Return(nil).Once()
	authMock.On("CleanupStalePasswordResetTokens").Return(nil).Once()
	authMock.On("CleanupStaleEmailVerificationTokens").Return(nil).Once()
	invMock.On("CleanupStale").Return(nil).Once()
	apiKeyMock.On("CleanupStaleExpired").Return(nil).Once()

	srvs := &services.Services{
		Organization:     orgMock,
		UserRefreshToken: urtMock,
		Auth:             authMock,
		Invitation:       invMock,
		ApiKey:           apiKeyMock,
	}
	jc := &Client{log: yca_log.New()}
	c := NewConsumers(srvs, jc, yca_log.New())

	err := c.cleanupHandler()()
	s.Require().NoError(err)

	orgMock.AssertExpectations(s.T())
}

func (s *CleanupHandlerSuite) TestIsNoRowsAffectedError() {
	s.True(isNoRowsAffectedError(errors.New("no rows affected")))
	s.True(isNoRowsAffectedError(yca_error.NewNotFoundError(errors.New("no rows affected"), "", nil)))
	s.False(isNoRowsAffectedError(errors.New("boom")))
	s.False(isNoRowsAffectedError(nil))
}
*/
