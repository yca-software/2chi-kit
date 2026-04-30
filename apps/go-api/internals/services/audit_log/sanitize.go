package audit_log_service

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
)

var emailLike = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// ToPublicAuditLog returns a client-safe copy with masked actor fields and sanitized data JSON.
func ToPublicAuditLog(log *models.AuditLog) models.AuditLogPublic {
	var data json.RawMessage
	if log.Data != nil {
		data = SanitizeAuditDataJSON(*log.Data)
	} else {
		data = json.RawMessage(`null`)
	}
	var impID *uuid.UUID
	impEmail := ""
	if log.ImpersonatedByID.Valid {
		u := log.ImpersonatedByID.UUID
		impID = &u
		impEmail = log.ImpersonatedByEmail
	}
	return models.AuditLogPublic{
		ID:                  log.ID,
		CreatedAt:           log.CreatedAt,
		OrganizationID:      log.OrganizationID,
		ActorID:             log.ActorID,
		ActorInfo:           log.ActorInfo,
		ImpersonatedByID:    impID,
		ImpersonatedByEmail: impEmail,
		Action:              log.Action,
		ResourceType:        log.ResourceType,
		ResourceID:          log.ResourceID,
		ResourceName:        log.ResourceName,
		Data:                data,
	}
}

// MaskActorDisplay masks email-shaped values; leaves API key prefixes and other labels as-is.
func MaskActorDisplay(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if emailLike.MatchString(s) {
		return maskEmail(s)
	}
	return s
}

func maskEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 0 || at >= len(email)-1 {
		return "***"
	}
	local, domain := email[:at], email[at+1:]
	if len(local) == 0 {
		return "***@" + domain
	}
	return string(local[0]) + "***@" + domain
}

func maskStringIfEmail(s string) string {
	s = strings.TrimSpace(s)
	if s != "" && emailLike.MatchString(s) {
		return maskEmail(s)
	}
	return s
}

// SanitizeAuditDataJSON redacts sensitive keys and masks email-shaped string values (safe for storage and API).
func SanitizeAuditDataJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage(`null`)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return json.RawMessage(`{}`)
	}
	sanitized := sanitizeValue(v, "")
	out, err := json.Marshal(sanitized)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return out
}

func sanitizeValue(v any, keyHint string) any {
	switch x := v.(type) {
	case map[string]any:
		return sanitizeMap(x)
	case []any:
		out := make([]any, len(x))
		for i, el := range x {
			out[i] = sanitizeValue(el, keyHint)
		}
		return out
	case string:
		if isSensitiveKey(keyHint) {
			return "[redacted]"
		}
		return maskStringIfEmail(x)
	case json.Number:
		if isSensitiveKey(keyHint) {
			return "[redacted]"
		}
		return x
	default:
		if isSensitiveKey(keyHint) {
			return "[redacted]"
		}
		return x
	}
}

func sanitizeMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			out[k] = "[redacted]"
			continue
		}
		switch child := v.(type) {
		case map[string]any:
			out[k] = sanitizeMap(child)
		case []any:
			out[k] = sanitizeValue(child, k)
		default:
			out[k] = sanitizeValue(v, k)
		}
	}
	return out
}

func isSensitiveKey(k string) bool {
	kl := strings.ToLower(k)
	switch kl {
	case "password", "currentpassword", "newpassword", "token", "secrettoken", "accesstoken",
		"refreshtoken", "authorization", "cookie", "apisecret", "clientsecret", "webhooksecret",
		"privatekey", "keyhash", "key_hash", "paddlecustomerid", "billingemail",
		"useremail", "memberemail", "invitedemail", "invitedbyemail":
		return true
	default:
		return strings.Contains(kl, "password") ||
			strings.Contains(kl, "secret") ||
			strings.Contains(kl, "token") ||
			kl == "api_key" ||
			strings.HasSuffix(kl, "apikey")
	}
}
