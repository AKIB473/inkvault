package apierr

import (
	"net/http"
	"testing"
)

func TestAPIErrorInterface(t *testing.T) {
	err := ErrPostNotFound
	if err.Error() != err.Message {
		t.Error("Error() should return Message")
	}
}

func TestWithMessage(t *testing.T) {
	custom := ErrForbidden.WithMessage("you need editor role")
	if custom.Message != "you need editor role" {
		t.Errorf("WithMessage: got %q", custom.Message)
	}
	// Should not mutate original
	if ErrForbidden.Message == "you need editor role" {
		t.Error("WithMessage mutated original error")
	}
}

func TestErrorStatusCodes(t *testing.T) {
	cases := []struct {
		err    APIError
		status int
	}{
		{ErrBadJSON, http.StatusBadRequest},
		{ErrUnauthorized, http.StatusUnauthorized},
		{ErrForbidden, http.StatusForbidden},
		{ErrPostNotFound, http.StatusNotFound},
		{ErrUsernameTaken, http.StatusConflict},
		{ErrPostUnpublished, http.StatusGone},
		{ErrRateLimited, http.StatusTooManyRequests},
		{ErrInternal, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		if tc.err.Status != tc.status {
			t.Errorf("%s: status = %d, want %d", tc.err.Code, tc.err.Status, tc.status)
		}
	}
}
