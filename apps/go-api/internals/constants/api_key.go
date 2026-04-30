package constants

// API key format and validation constants.
//
// API_KEY_PREFIX ("sk_") is required on all API keys. It is:
//   - Prepended when creating keys (api_key/create.go)
//   - Stripped before hashing in PrivateRoute middleware (X-API-Key header)
//   - Used consistently: clients must send "sk_<raw_key>"; the raw part is hashed for lookup
//
// Changing the prefix would invalidate existing keys. Keep it consistent across create, auth, and display.
const (
	API_KEY_PREFIX     = "sk_"
	API_KEY_PREFIX_LEN = 8 // characters after prefix for display (e.g. sk_a1b2c3d4)
	API_KEY_RANDOM_LEN = 32
)
