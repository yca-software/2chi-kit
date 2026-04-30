package jobs

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	paddle_service "github.com/yca-software/2chi-kit/go-api/internals/services/paddle"
	yca_log "github.com/yca-software/go-common/logger"
)

type ApplyScheduledPlanChangesHandlerSuite struct {
	suite.Suite
}

func TestApplyScheduledPlanChangesHandlerSuite(t *testing.T) {
	suite.Run(t, new(ApplyScheduledPlanChangesHandlerSuite))
}

func (s *ApplyScheduledPlanChangesHandlerSuite) TestDelegatesToPaddle() {
	paddleMock := paddle_service.NewMockPaddleService()
	ctx := context.Background()
	paddleMock.On("ApplyScheduledPlanChanges", mock.MatchedBy(func(c context.Context) bool { return c == ctx })).Return(nil).Once()

	srvs := &services.Services{Paddle: paddleMock}
	jc := &Client{log: yca_log.New()}
	c := NewConsumers(srvs, jc, yca_log.New())

	err := c.applyScheduledPlanChangesHandler(ctx)()
	s.Require().NoError(err)
	paddleMock.AssertExpectations(s.T())
}

func (s *ApplyScheduledPlanChangesHandlerSuite) TestPropagatesError() {
	paddleMock := paddle_service.NewMockPaddleService()
	ctx := context.Background()
	paddleMock.On("ApplyScheduledPlanChanges", mock.Anything).Return(errors.New("paddle failed")).Once()

	srvs := &services.Services{Paddle: paddleMock}
	jc := &Client{log: yca_log.New()}
	c := NewConsumers(srvs, jc, yca_log.New())

	err := c.applyScheduledPlanChangesHandler(ctx)()
	s.Require().Error(err)
	paddleMock.AssertExpectations(s.T())
}
