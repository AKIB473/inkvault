package crypto

import (
	"testing"
	"time"
)

func TestGenerateSecureToken(t *testing.T) {
	tok1, err := GenerateSecureToken(32)
	if err != nil {
		t.Fatalf("GenerateSecureToken: %v", err)
	}
	// 32 bytes = 64 hex chars
	if len(tok1) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(tok1))
	}

	// Uniqueness
	tok2, _ := GenerateSecureToken(32)
	if tok1 == tok2 {
		t.Error("two tokens are identical — RNG may be broken")
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	h1 := HashToken("my-token")
	h2 := HashToken("my-token")
	if h1 != h2 {
		t.Error("HashToken not deterministic")
	}
	if len(h1) != 64 {
		t.Errorf("expected 64-char SHA-256 hex, got %d", len(h1))
	}
}

func TestHashTokenUnique(t *testing.T) {
	if HashToken("token-a") == HashToken("token-b") {
		t.Error("different tokens produced same hash")
	}
}

func TestSignToken(t *testing.T) {
	signed := SignToken("mytoken", time.Now().Add(time.Hour), []byte("secret"))
	if signed == "" {
		t.Error("SignToken returned empty string")
	}
	// Format: <token>.<expiry>.<signature> — must have at least 2 dots
	dots := 0
	for _, c := range signed {
		if c == '.' {
			dots++
		}
	}
	if dots < 2 {
		t.Errorf("SignToken output missing dots: %q", signed)
	}
	// Different secrets produce different signatures
	signed2 := SignToken("mytoken", time.Now().Add(time.Hour), []byte("othersecret"))
	if signed == signed2 {
		t.Error("different secrets produced same signature")
	}
}

func TestGenerateSpecialTokens(t *testing.T) {
	invite, err := GenerateInviteToken()
	if err != nil {
		t.Fatalf("GenerateInviteToken: %v", err)
	}
	if len(invite) < 32 {
		t.Error("invite token too short")
	}

	reset, err := GeneratePasswordResetToken()
	if err != nil {
		t.Fatalf("GeneratePasswordResetToken: %v", err)
	}
	if invite == reset {
		t.Error("tokens are identical across calls")
	}
}
