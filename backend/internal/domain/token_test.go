package domain

import (
	"testing"
	"time"
)

func TestTokenIsValid(t *testing.T) {
	now := time.Now()

	valid := &Token{
		ExpiresAt: now.Add(1 * time.Hour),
		UsedAt:    nil,
	}
	if !valid.IsValid() {
		t.Error("unexpired unused token should be valid")
	}

	expired := &Token{
		ExpiresAt: now.Add(-1 * time.Second),
		UsedAt:    nil,
	}
	if expired.IsValid() {
		t.Error("expired token should be invalid")
	}

	used := &Token{
		ExpiresAt: now.Add(1 * time.Hour),
		UsedAt:    &now,
	}
	if used.IsValid() {
		t.Error("used token should be invalid")
	}

	expiredAndUsed := &Token{
		ExpiresAt: now.Add(-1 * time.Hour),
		UsedAt:    &now,
	}
	if expiredAndUsed.IsValid() {
		t.Error("expired+used token should be invalid")
	}
}
