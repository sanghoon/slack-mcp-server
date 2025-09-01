package utils

import (
	"errors"
	"testing"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "slack bot token",
			input:    "Using token xoxb-1234567890-abcdefghijklmnop",
			expected: "Using token xoxb-12[REDACTED]",
		},
		{
			name:     "slack user token",
			input:    "Auth with xoxp-987654321-zyxwvutsrqp",
			expected: "Auth with xoxp-98[REDACTED]",
		},
		{
			name:     "multiple tokens",
			input:    "Tokens: xoxc-aaa-bbb and xoxd-ccc-ddd",
			expected: "Tokens: xoxc-aa[REDACTED] and xoxd-cc[REDACTED]",
		},
		{
			name:     "bearer token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "Authorization: Bearer [REDACTED]",
		},
		{
			name:     "api key",
			input:    "api_key: sk-1234567890abcdef",
			expected: "api_key: [REDACTED]",
		},
		{
			name:     "no tokens",
			input:    "This is a regular message without tokens",
			expected: "This is a regular message without tokens",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "error with token",
			err:      errors.New("authentication failed with token xoxb-secret-token"),
			expected: "authentication failed with token xoxb-se[REDACTED]",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "regular error",
			err:      errors.New("connection timeout"),
			expected: "connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := SanitizeError(tt.err)
			if tt.err == nil && sanitized != nil {
				t.Errorf("SanitizeError(nil) = %v, want nil", sanitized)
			} else if tt.err != nil && sanitized.Error() != tt.expected {
				t.Errorf("SanitizeError() = %q, want %q", sanitized.Error(), tt.expected)
			}
		})
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "bot token",
			token:    "xoxb-1234567890",
			expected: "xoxb-****",
		},
		{
			name:     "user token",
			token:    "xoxp-abcdefghij",
			expected: "xoxp-****",
		},
		{
			name:     "client token",
			token:    "xoxc-zyxwvutsrq",
			expected: "xoxc-****",
		},
		{
			name:     "cookie token",
			token:    "xoxd-qrstuvwxyz",
			expected: "xoxd-****",
		},
		{
			name:     "unknown token",
			token:    "sk-1234567890",
			expected: "sk-1****",
		},
		{
			name:     "short token",
			token:    "abc",
			expected: "****",
		},
		{
			name:     "empty token",
			token:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskToken(tt.token)
			if result != tt.expected {
				t.Errorf("MaskToken() = %q, want %q", result, tt.expected)
			}
		})
	}
}
