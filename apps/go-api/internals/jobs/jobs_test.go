package jobs

import (
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	yca_error "github.com/yca-software/go-common/error"
)

type JobHelpersSuite struct {
	suite.Suite
}

func TestJobHelpersSuite(t *testing.T) {
	suite.Run(t, new(JobHelpersSuite))
}

func (s *JobHelpersSuite) TestParseRetryCount() {
	s.Equal(0, parseRetryCount(nil))
	s.Equal(0, parseRetryCount(amqp.Table{}))
	s.Equal(2, parseRetryCount(amqp.Table{HeaderRetryCount: int32(2)}))
	s.Equal(3, parseRetryCount(amqp.Table{HeaderRetryCount: int64(3)}))
}

func (s *JobHelpersSuite) TestIsRetryable_UnknownFalse() {
	s.False(IsRetryable(errors.New("business logic")))
}

func (s *JobHelpersSuite) TestIsRetryable_RetryableWrapper() {
	s.True(IsRetryable(Retryable(errors.New("wrapped"))))
}

func (s *JobHelpersSuite) TestIsRetryable_ycaError_5xxOnly() {
	s.True(IsRetryable(yca_error.NewInternalServerError(errors.New("db"), "", nil)))
	s.False(IsRetryable(yca_error.NewNotFoundError(errors.New("missing"), "NOT_FOUND", nil)))
}

func (s *JobHelpersSuite) TestClassifyJobError_NonRetryableDeadLetter() {
	dl, rp := classifyJobError(errors.New("bad input"), 0, 3)
	s.True(dl)
	s.False(rp)
}

func (s *JobHelpersSuite) TestClassifyJobError_RetryableUnderCap() {
	dl, rp := classifyJobError(Retryable(errors.New("db down")), 0, 3)
	s.False(dl)
	s.True(rp)
}

func (s *JobHelpersSuite) TestClassifyJobError_RetryableExhausted() {
	dl, rp := classifyJobError(Retryable(errors.New("db down")), 3, 3)
	s.True(dl)
	s.False(rp)
}
