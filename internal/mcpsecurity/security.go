package mcpsecurity

import "crypto/subtle"

const AuthTokenMetadataKey = "x-mcp-token"

func TokenMatches(expected, actual string) bool {
	if expected == "" {
		return true
	}
	if actual == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
