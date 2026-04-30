package jobs

import (
	"errors"
	"net/http"

	yca_error "github.com/yca-software/go-common/error"
)

// errRetryableMarker wraps an error to force IsRetryable to return true.
type errRetryableMarker struct{ err error }

func (e *errRetryableMarker) Error() string { return e.err.Error() }
func (e *errRetryableMarker) Unwrap() error { return e.err }

// Retryable wraps err so job processing treats it as infrastructure-retryable
// (counts against JOB_INFRA_MAX_RETRIES before dead-letter).
func Retryable(err error) error {
	if err == nil {
		return nil
	}
	return &errRetryableMarker{err: err}
}

// IsRetryable reports whether err should be retried (republish up to JOB_INFRA_MAX_RETRIES).
// True for: explicit Retryable() wrapper, or *yca_error.Error with HTTP status 5xx only.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := errors.AsType[*errRetryableMarker](err); ok {
		return true
	}
	var apiErr *yca_error.Error
	if errors.As(err, &apiErr) && apiErr != nil {
		code := apiErr.StatusCode
		return code >= http.StatusInternalServerError && code <= 599
	}
	return false
}
