package users

import (
	"testing"

	"github.com/you/inkvault/internal/domain"
)

func TestCanPerform(t *testing.T) {
	owner := &domain.User{Role: domain.RoleOwner, Status: domain.UserStatusActive}
	admin := &domain.User{ID: "admin1", Role: domain.RoleAdmin, Status: domain.UserStatusActive}
	editor := &domain.User{ID: "editor1", Role: domain.RoleEditor, Status: domain.UserStatusActive}
	writer := &domain.User{ID: "writer1", Role: domain.RoleWriter, Status: domain.UserStatusActive}
	reader := &domain.User{Role: domain.RoleReader, Status: domain.UserStatusActive}
	banned := &domain.User{Role: domain.RoleAdmin, Status: domain.UserStatusBanned}

	cases := []struct {
		label   string
		user    *domain.User
		action  Action
		res     ResourceType
		ownerID string
		want    bool
	}{
		// Owner: can do anything
		{"owner can delete user", owner, ActionDelete, ResourceUser, "", true},
		{"owner can create post", owner, ActionCreate, ResourcePost, "", true},

		// Admin: can do most things
		{"admin can edit post", admin, ActionEdit, ResourcePost, "", true},
		{"admin can delete post", admin, ActionDelete, ResourcePost, "", true},

		// Editor: can manage all posts
		{"editor can edit all posts", editor, ActionEdit, ResourcePost, "", true},
		{"editor can delete post", editor, ActionDelete, ResourcePost, "", true},
		{"editor cannot manage users", editor, ActionDelete, ResourceUser, "", false},

		// Writer: can only manage own content
		{"writer can create post", writer, ActionCreate, ResourcePost, "", true},
		{"writer can edit own post", writer, ActionEdit, ResourcePost, "writer1", true},
		{"writer cannot edit others post", writer, ActionEdit, ResourcePost, "other", false},
		{"writer cannot manage settings", writer, ActionEdit, ResourceSettings, "", false},

		// Reader: read only
		{"reader can read", reader, ActionRead, ResourcePost, "", true},
		{"reader cannot create", reader, ActionCreate, ResourcePost, "", false},
		{"reader cannot delete", reader, ActionDelete, ResourcePost, "", false},

		// Banned: nothing
		{"banned user cannot do anything", banned, ActionRead, ResourcePost, "", false},

		// Nil user
		{"nil user cannot do anything", nil, ActionRead, ResourcePost, "", false},
	}

	for _, tc := range cases {
		got := CanPerform(tc.user, tc.action, tc.res, tc.ownerID)
		if got != tc.want {
			t.Errorf("[%s] CanPerform() = %v, want %v", tc.label, got, tc.want)
		}
	}
}

func TestMagicLinkConfig(t *testing.T) {
	cfg := MagicLinkConfig()
	if cfg["validity_seconds"] != MagicLinkTokenValidity {
		t.Error("wrong validity")
	}
	if cfg["max_usage_count"] != MagicLinkTokenMaxUsageCount {
		t.Error("wrong max_usage_count")
	}
	// 10min post-use window (Ghost: prevents email client pre-fetch from consuming token)
	if cfg["validity_after_usage_seconds"] != MagicLinkTokenValidityAfterUsage {
		t.Error("wrong validity_after_usage_seconds")
	}
}
