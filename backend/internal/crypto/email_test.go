package crypto

import (
	"strings"
	"testing"
)

const testKey = "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

func TestEmailEncryptorRoundTrip(t *testing.T) {
	enc, err := NewEmailEncryptor(testKey)
	if err != nil {
		t.Fatalf("NewEmailEncryptor: %v", err)
	}

	emails := []string{
		"user@example.com",
		"complex+tag@sub.domain.org",
		"unicode@münchen.de",
		"", // edge: empty
	}

	for _, email := range emails {
		if email == "" {
			continue
		}
		cipher, err := enc.Encrypt(email)
		if err != nil {
			t.Errorf("Encrypt(%q): %v", email, err)
			continue
		}
		if cipher == email {
			t.Errorf("Encrypt(%q): ciphertext equals plaintext", email)
		}

		plain, err := enc.Decrypt(cipher)
		if err != nil {
			t.Errorf("Decrypt(%q): %v", email, err)
			continue
		}
		if plain != email {
			t.Errorf("roundtrip: got %q, want %q", plain, email)
		}
	}
}

func TestEmailEncryptorNonDeterministic(t *testing.T) {
	enc, _ := NewEmailEncryptor(testKey)
	email := "same@email.com"

	c1, _ := enc.Encrypt(email)
	c2, _ := enc.Encrypt(email)
	// Two encryptions of same plaintext must produce different ciphertext (random nonce)
	if c1 == c2 {
		t.Error("two encryptions of same email produced identical ciphertext — nonce not random")
	}
}

func TestEmailEncryptorBadKey(t *testing.T) {
	// Short key
	if _, err := NewEmailEncryptor("deadbeef"); err == nil {
		t.Error("expected error for short key")
	}
	// Invalid hex
	if _, err := NewEmailEncryptor(strings.Repeat("zz", 32)); err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestEmailEncryptorTampering(t *testing.T) {
	enc, _ := NewEmailEncryptor(testKey)
	cipher, _ := enc.Encrypt("legit@example.com")

	// Flip a byte in the ciphertext — GCM authentication must fail
	tampered := []byte(cipher)
	tampered[len(tampered)/2] ^= 0xFF
	if _, err := enc.Decrypt(string(tampered)); err == nil {
		t.Error("expected error when decrypting tampered ciphertext")
	}
}
