package jobs

import (
	"strings"
	"sync"

	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type cleanupStep struct {
	name string
	run  func() error
}

func (c *Consumers) cleanupHandler() func() error {
	steps := []cleanupStep{
		{name: "organization.CleanupArchived", run: c.srvs.Organization.CleanupArchived},
		{name: "user_refresh_token.CleanupStaleUnused", run: c.srvs.UserRefreshToken.CleanupStaleUnused},
		{name: "auth.CleanupStalePasswordResetTokens", run: c.srvs.Auth.CleanupStalePasswordResetTokens},
		{name: "auth.CleanupStaleEmailVerificationTokens", run: c.srvs.Auth.CleanupStaleEmailVerificationTokens},
		{name: "invitation.CleanupStale", run: c.srvs.Invitation.CleanupStale},
		{name: "api_key.CleanupStaleExpired", run: c.srvs.ApiKey.CleanupStaleExpired},
	}
	return func() error {
		var wg sync.WaitGroup
		for _, step := range steps {
			wg.Add(1)
			go func(s cleanupStep) {
				defer wg.Done()
				if err := s.run(); err != nil {
					if isNoRowsAffectedError(err) {
						return
					}
					c.logger.Log(yca_log.LogData{
						Level:    "error",
						Location: "jobs.cleanupHandler",
						Error:    err,
						Message:  "cleanup step failed",
						Data:     map[string]any{"step": s.name},
					})
				}
			}(step)
		}
		wg.Wait()
		return nil
	}
}

func isNoRowsAffectedError(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(strings.ToLower(err.Error()), "no rows affected") {
		return true
	}
	if apiErr, ok := yca_error.AsError(err); ok && apiErr.Err != nil {
		return strings.Contains(strings.ToLower(apiErr.Err.Error()), "no rows affected")
	}
	return false
}
