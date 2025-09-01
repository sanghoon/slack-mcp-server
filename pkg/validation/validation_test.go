package validation

import (
	"strings"
	"testing"
)

func TestValidateChannelID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid channel ID", "C1234567890", false},
		{"valid DM ID", "D1234567890", false},
		{"channel name with #", "#general", false},
		{"channel name with @", "@username", false},
		{"empty input", "", true},
		{"invalid format", "X1234567890", true},
		{"too short", "C123", true},
		{"lowercase", "c1234567890", true},
		{"special characters", "C123!@#$%", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChannelID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateChannelID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUserID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid user ID", "U1234567890", false},
		{"valid workspace ID", "W1234567890", false},
		{"empty input", "", true},
		{"invalid prefix", "X1234567890", true},
		{"too short", "U123", true},
		{"lowercase", "u1234567890", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid timestamp", "1234567890.123456", false},
		{"empty (optional)", "", false},
		{"missing decimal", "1234567890", true},
		{"wrong decimal places", "1234567890.123", true},
		{"non-numeric", "abc.123456", true},
		{"too long", "12345678901.123456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimestamp(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimestamp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSearchQuery(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple query", "hello world", "hello world", false},
		{"query with quotes", `test "quoted text"`, `test \"quoted text\"`, false},
		{"query with special chars", "test <script>alert()</script>", `test \<script\>alert\(\)\</script\>`, false},
		{"empty query", "", "", true},
		{"too long query", strings.Repeat("a", 1001), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateSearchQuery(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSearchQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateSearchQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeSearchQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no special chars", "hello world", "hello world"},
		{"with quotes", `hello "world"`, `hello \"world\"`},
		{"with backslash", `path\to\file`, `path\\to\\file`},
		{"with brackets", "array[0]", `array\[0\]`},
		{"with wildcards", "test*", `test\*`},
		{"complex query", `<script>alert("xss")</script>`, `\<script\>alert\(\"xss\"\)\</script\>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeSearchQuery(tt.input); got != tt.want {
				t.Errorf("SanitizeSearchQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid message", "Hello, world!", false},
		{"empty message", "", true},
		{"max length", strings.Repeat("a", 40000), false},
		{"too long", strings.Repeat("a", 40001), true},
		{"with emojis", "Hello 👋 World 🌍", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLimit(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid limit", 100, false},
		{"min limit", 1, false},
		{"max limit", 1000, false},
		{"zero", 0, true},
		{"negative", -1, true},
		{"too large", 1001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLimit(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCursor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid base64", "dGVhbTpDMDYxRTAxUUE=", false},
		{"empty (optional)", "", false},
		{"with padding", "SGVsbG8gV29ybGQ=", false},
		{"invalid chars", "not-base64!", true},
		{"too long", strings.Repeat("A", 201), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCursor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCursor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
