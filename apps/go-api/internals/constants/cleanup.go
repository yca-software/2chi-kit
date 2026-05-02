package constants

import "time"

// StaleDataRetentionPeriod is how long rows are kept after they are no longer in use before the cleanup job hard-deletes them.
var StaleDataRetentionPeriod = 30 * 24 * time.Hour

// InvitationPendingExpiredCleanupMinAge is how long after expires_at a still-pending (never accepted or revoked)
// invitation may be hard-deleted. Kept longer than StaleDataRetentionPeriod so auth can still resolve the row
// and return INVITATION_EXPIRED instead of treating a missing row as INVALID_INVITATION_TOKEN.
var InvitationPendingExpiredCleanupMinAge = 90 * 24 * time.Hour

// ArchivedSoftDeleteRetentionPeriod is how long soft-deleted organizations, warehouses, and HoReCa accounts remain before the cleanup job permanently removes them.
var ArchivedSoftDeleteRetentionPeriod = 90 * 24 * time.Hour
