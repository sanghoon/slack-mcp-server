package utils

import (
	"regexp"
	"strings"
)

var (
	// Slack token patterns
	slackTokenRegex = regexp.MustCompile(`xox[pbcd]-[a-zA-Z0-9\-]+`)
	// Generic token patterns (Bearer tokens, API keys, etc.)
	bearerTokenRegex = regexp.MustCompile(`Bearer\s+[a-zA-Z0-9\-._~+/]+=*`)
	apiKeyRegex      = regexp.MustCompile(`[aA][pP][iI][-_]?[kK][eE][yY]\s*[:=]\s*[a-zA-Z0-9\-._~+/]+=*`)
)

// SanitizeString removes sensitive tokens from a string
func SanitizeString(input string) string {
	if input == "" {
		return input
	}

	// Replace Slack tokens
	sanitized := slackTokenRegex.ReplaceAllStringFunc(input, func(match string) string {
		if len(match) > 10 {
			return match[:7] + "[REDACTED]"
		}
		return "[REDACTED]"
	})

	// Replace Bearer tokens
	sanitized = bearerTokenRegex.ReplaceAllString(sanitized, "Bearer [REDACTED]")

	// Replace API keys
	sanitized = apiKeyRegex.ReplaceAllString(sanitized, "api_key: [REDACTED]")

	return sanitized
}

// SanitizeError creates a new error with sanitized message
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}
	return &sanitizedError{
		original: err,
		message:  SanitizeString(err.Error()),
	}
}

type sanitizedError struct {
	original error
	message  string
}

func (e *sanitizedError) Error() string {
	return e.message
}

func (e *sanitizedError) Unwrap() error {
	return e.original
}

// MaskToken masks a token for display purposes (e.g., in logs)
func MaskToken(token string) string {
	if token == "" {
		return ""
	}

	// Determine token type and mask accordingly
	switch {
	case strings.HasPrefix(token, "xoxb-"):
		return "xoxb-****"
	case strings.HasPrefix(token, "xoxp-"):
		return "xoxp-****"
	case strings.HasPrefix(token, "xoxc-"):
		return "xoxc-****"
	case strings.HasPrefix(token, "xoxd-"):
		return "xoxd-****"
	default:
		if len(token) > 8 {
			return token[:4] + "****"
		}
		return "****"
	}
}
