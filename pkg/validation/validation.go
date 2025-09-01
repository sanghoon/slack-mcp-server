package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	// Slack ID patterns
	channelIDRegex = regexp.MustCompile(`^[CD][A-Z0-9]{8,}$`)
	userIDRegex    = regexp.MustCompile(`^[UW][A-Z0-9]{8,}$`)

	// Channel name patterns (without prefix)
	channelNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

	// Timestamp pattern
	timestampRegex = regexp.MustCompile(`^\d{10}\.\d{6}$`)

	// Maximum lengths
	maxChannelNameLength = 80
	maxSearchQueryLength = 1000
	maxMessageLength     = 40000
	maxLimitValue        = 1000
)

// Errors
var (
	ErrInvalidChannelID   = errors.New("invalid channel ID format")
	ErrInvalidChannelName = errors.New("invalid channel name format")
	ErrInvalidUserID      = errors.New("invalid user ID format")
	ErrInvalidTimestamp   = errors.New("invalid timestamp format")
	ErrQueryTooLong       = errors.New("search query too long")
	ErrMessageTooLong     = errors.New("message too long")
	ErrInvalidLimit       = errors.New("limit must be between 1 and 1000")
	ErrEmptyInput         = errors.New("input cannot be empty")
)

// ValidateChannelID validates a Slack channel ID
func ValidateChannelID(channelID string) error {
	if channelID == "" {
		return ErrEmptyInput
	}

	// Remove # or @ prefix if present
	if strings.HasPrefix(channelID, "#") || strings.HasPrefix(channelID, "@") {
		return ValidateChannelName(channelID)
	}

	if !channelIDRegex.MatchString(channelID) {
		return fmt.Errorf("%w: %s", ErrInvalidChannelID, channelID)
	}

	return nil
}

// ValidateChannelName validates a channel name (with or without prefix)
func ValidateChannelName(name string) error {
	if name == "" {
		return ErrEmptyInput
	}

	// Remove prefix if present
	cleanName := strings.TrimPrefix(strings.TrimPrefix(name, "#"), "@")

	if len(cleanName) > maxChannelNameLength {
		return fmt.Errorf("channel name too long: max %d characters", maxChannelNameLength)
	}

	if !channelNameRegex.MatchString(cleanName) {
		return fmt.Errorf("%w: %s", ErrInvalidChannelName, cleanName)
	}

	return nil
}

// ValidateUserID validates a Slack user ID
func ValidateUserID(userID string) error {
	if userID == "" {
		return ErrEmptyInput
	}

	if !userIDRegex.MatchString(userID) {
		return fmt.Errorf("%w: %s", ErrInvalidUserID, userID)
	}

	return nil
}

// ValidateTimestamp validates a Slack timestamp
func ValidateTimestamp(ts string) error {
	if ts == "" {
		return nil // Empty timestamp is valid (optional parameter)
	}

	if !timestampRegex.MatchString(ts) {
		return fmt.Errorf("%w: %s", ErrInvalidTimestamp, ts)
	}

	return nil
}

// ValidateSearchQuery validates and sanitizes a search query
func ValidateSearchQuery(query string) (string, error) {
	if query == "" {
		return "", ErrEmptyInput
	}

	// Check length
	if utf8.RuneCountInString(query) > maxSearchQueryLength {
		return "", fmt.Errorf("%w: max %d characters", ErrQueryTooLong, maxSearchQueryLength)
	}

	// Sanitize the query
	sanitized := SanitizeSearchQuery(query)

	return sanitized, nil
}

// SanitizeSearchQuery removes potentially dangerous characters from search queries
func SanitizeSearchQuery(query string) string {
	// Escape special characters that have meaning in Slack search
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		`'`, `\'`,
		`<`, `\<`,
		`>`, `\>`,
		`|`, `\|`,
		`&`, `\&`,
		`;`, `\;`,
		`$`, `\$`,
		"`", "\\`",
		`!`, `\!`,
		`*`, `\*`,
		`?`, `\?`,
		`[`, `\[`,
		`]`, `\]`,
		`(`, `\(`,
		`)`, `\)`,
		`{`, `\{`,
		`}`, `\}`,
	)

	return replacer.Replace(query)
}

// ValidateMessage validates message content
func ValidateMessage(message string) error {
	if message == "" {
		return ErrEmptyInput
	}

	if utf8.RuneCountInString(message) > maxMessageLength {
		return fmt.Errorf("%w: max %d characters", ErrMessageTooLong, maxMessageLength)
	}

	return nil
}

// ValidateLimit validates a limit parameter
func ValidateLimit(limit int) error {
	if limit < 1 || limit > maxLimitValue {
		return fmt.Errorf("%w: %d", ErrInvalidLimit, limit)
	}
	return nil
}

// ValidateCursor validates a cursor for pagination
func ValidateCursor(cursor string) error {
	if cursor == "" {
		return nil // Empty cursor is valid
	}

	// Basic length check to prevent extremely long cursors
	if len(cursor) > 200 {
		return errors.New("cursor too long")
	}

	// Cursor should only contain base64 characters
	base64Regex := regexp.MustCompile(`^[A-Za-z0-9+/=]+$`)
	if !base64Regex.MatchString(cursor) {
		return errors.New("invalid cursor format")
	}

	return nil
}
