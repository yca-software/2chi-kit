package support_service

import (
	"fmt"
	"html"
	"strings"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type SubmitRequest struct {
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	PageURL   string `json:"pageUrl,omitempty" validate:"omitempty,url"`
	UserAgent string `json:"userAgent,omitempty" validate:"omitempty"`
}

func (s *service) Submit(req *SubmitRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}
	if accessInfo == nil {
		return yca_error.NewUnauthorizedError(nil, constants.UNAUTHORIZED_CODE, nil)
	}
	if accessInfo.User == nil {
		return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
	}
	if s.supportInboxEmail == "" {
		return yca_error.NewInternalServerError(nil, constants.INTERNAL_SERVER_ERROR_CODE, nil)
	}

	subject := strings.TrimSpace(req.Subject)
	if subject == "" {
		subject = fmt.Sprintf("Support request from %s", accessInfo.User.Email)
	} else {
		subject = "[Support] " + subject
	}

	body := fmt.Sprintf(
		`<p><strong>From</strong>: %s</p>
<p><strong>User ID</strong>: %s</p>
<p><strong>Subject</strong>: %s</p>
<p><strong>Message</strong>:</p>
<pre style="white-space:pre-wrap;font-family:inherit;">%s</pre>
<p><strong>Page URL</strong>: %s</p>
<p><strong>Client user agent</strong>: %s</p>
<p><strong>Request IP</strong>: %s</p>`,
		html.EscapeString(accessInfo.User.Email),
		html.EscapeString(accessInfo.User.UserID.String()),
		html.EscapeString(strings.TrimSpace(req.Subject)),
		html.EscapeString(req.Message),
		html.EscapeString(req.PageURL),
		html.EscapeString(req.UserAgent),
		html.EscapeString(accessInfo.IPAddress),
	)

	if err := s.emailService.SendEmail(s.supportInboxEmail, subject, body); err != nil {
		return err
	}
	return nil
}
