package user

import "strings"

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	lowerErr := strings.ToLower(err.Error())
	return strings.Contains(lowerErr, "duplicate key") || strings.Contains(lowerErr, "unique")
}
