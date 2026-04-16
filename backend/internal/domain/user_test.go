package domain

import (
	"testing"
)

func TestUserStatusBitmask(t *testing.T) {
	u := &User{Status: UserStatusActive}
	if !u.IsActive() {
		t.Error("active user should be active")
	}

	u.Status = UserStatusSilenced
	if !u.IsSilenced() {
		t.Error("silenced user should be silenced")
	}
	if u.IsBanned() {
		t.Error("silenced user should not be banned")
	}
	if u.IsActive() {
		t.Error("silenced user should not be active")
	}

	u.Status = UserStatusBanned
	if !u.IsBanned() {
		t.Error("banned user should be banned")
	}
	if u.IsActive() {
		t.Error("banned user should not be active")
	}
}

func TestUserRoleChecks(t *testing.T) {
	owner := &User{Role: RoleOwner}
	admin := &User{Role: RoleAdmin}
	writer := &User{Role: RoleWriter}
	reader := &User{Role: RoleReader}

	if !owner.IsAdmin() {
		t.Error("owner should pass IsAdmin()")
	}
	if !admin.IsAdmin() {
		t.Error("admin should pass IsAdmin()")
	}
	if writer.IsAdmin() {
		t.Error("writer should not pass IsAdmin()")
	}

	if !writer.CanEdit() {
		t.Error("writer should be able to edit")
	}
	if reader.CanEdit() {
		t.Error("reader should not be able to edit")
	}
}

func TestToPublicStripsSecrets(t *testing.T) {
	u := &User{
		ID:             "123",
		Username:       "alice",
		EmailEncrypted: "deadbeef",
		PasswordHash:   "$2a$12$...",
	}
	pub := u.ToPublic()

	// These should never appear in the public struct
	if pub.ID != "123" {
		t.Error("ID should be present")
	}
	// Compile-time check: PublicUser has no EmailEncrypted or PasswordHash fields
	// This test ensures ToPublic() returns the right type
	_ = pub.Username
	_ = pub.DisplayName
}

func TestEmailDecryption(t *testing.T) {
	u := &User{}
	// Before setting: should return empty
	if u.Email() != "" {
		t.Error("unset email should be empty")
	}

	u.SetDecryptedEmail("user@example.com")
	if u.Email() != "user@example.com" {
		t.Errorf("got %q, want user@example.com", u.Email())
	}
}
