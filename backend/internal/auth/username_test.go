package auth

import (
	"testing"
)

func TestNormalizeUsername(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Alice", "alice"},
		{"john_doe", "john_doe"},
		{"UPPER123", "upper123"},
		{"  spaces  ", "spaces"},
		{"hello-world", "helloworld"},  // dashes stripped
		{"with.dots", "withdots"},
		{"ab", ""},   // too short (< 3)
		{"a", ""},
		{"", ""},
		{"valid_123", "valid_123"},
	}

	for _, tc := range cases {
		got := normalizeUsername(tc.input)
		if got != tc.want {
			t.Errorf("normalizeUsername(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestReservedUsernames(t *testing.T) {
	reserved := []string{"admin", "api", "dashboard", "me", "login", "register", "www"}
	for _, name := range reserved {
		if !reservedUsernames[name] {
			t.Errorf("%q should be reserved but isn't", name)
		}
	}
}

func TestNormalizeUsernameMaxLength(t *testing.T) {
	long := "abcdefghij" + "abcdefghij" + "abcdefghij" // 30 chars — exactly at limit
	got := normalizeUsername(long)
	if got != long {
		t.Errorf("30-char username should pass: got %q", got)
	}

	tooLong := long + "x" // 31 chars
	got2 := normalizeUsername(tooLong)
	if got2 != "" {
		t.Errorf("31-char username should fail regex, got %q", got2)
	}
}
