package constants

import "time"

// StaleDataRetentionPeriod is how long rows are kept after they are no longer in use before the cleanup job hard-deletes them.
var StaleDataRetentionPeriod = 30 * 24 * time.Hour

// ArchivedSoftDeleteRetentionPeriod is how long soft-deleted organizations, warehouses, and HoReCa accounts remain before the cleanup job permanently removes them.
var ArchivedSoftDeleteRetentionPeriod = 90 * 24 * time.Hour
