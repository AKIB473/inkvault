// Package apierr defines all typed API errors for consistent HTTP responses.
// Learned from WriteFreely's errors.go — typed sentinel errors with HTTP status codes
// ensure handlers always return predictable, well-formed error responses.
package apierr

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// APIError is a typed error with an HTTP status code and machine-readable code.
type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"error"`
	Message string `json:"message"`
}

func (e APIError) Error() string { return e.Message }

// FiberResponse sends this error as a JSON response.
func (e APIError) FiberResponse(c *fiber.Ctx) error {
	return c.Status(e.Status).JSON(e)
}

// --- 400 Bad Request ---
var (
	ErrBadJSON        = APIError{http.StatusBadRequest, "invalid_json", "Expected valid JSON body."}
	ErrBadFormData    = APIError{http.StatusBadRequest, "invalid_form", "Expected valid form data."}
	ErrValidation     = APIError{http.StatusBadRequest, "validation_error", "Request validation failed."}
	ErrNoContent      = APIError{http.StatusBadRequest, "no_content", "Supply some content to publish."}
	ErrNoUpdateValues = APIError{http.StatusBadRequest, "no_update_values", "Supply at least one field to update."}
)

// --- 401 Unauthorized ---
var (
	ErrUnauthorized        = APIError{http.StatusUnauthorized, "unauthorized", "Authentication required."}
	ErrInvalidToken        = APIError{http.StatusUnauthorized, "invalid_token", "Invalid or expired token."}
	ErrInvalidCredentials  = APIError{http.StatusUnauthorized, "invalid_credentials", "Invalid username or password."}
	ErrNoAccessToken       = APIError{http.StatusUnauthorized, "no_token", "Authorization token required."}
	ErrBadAccessToken      = APIError{http.StatusUnauthorized, "bad_token", "Invalid access token."}
	ErrDisabledPasswordAuth = APIError{http.StatusForbidden, "password_auth_disabled", "Password authentication is disabled on this instance."}
)

// --- 403 Forbidden ---
var (
	ErrForbidden            = APIError{http.StatusForbidden, "forbidden", "Insufficient permissions."}
	ErrForbiddenEdit        = APIError{http.StatusForbidden, "forbidden_edit", "You don't have permission to edit this content."}
	ErrForbiddenCollection  = APIError{http.StatusForbidden, "forbidden_collection", "You don't have permission to access this blog."}
	ErrAccountBanned        = APIError{http.StatusForbidden, "account_banned", "Your account has been suspended."}
	ErrAccountSilenced      = APIError{http.StatusForbidden, "account_silenced", "Your account is silenced."}
	ErrRegistrationClosed   = APIError{http.StatusForbidden, "registration_closed", "Registration is currently closed."}
	ErrInvalidInvite        = APIError{http.StatusForbidden, "invalid_invite", "Invalid or expired invite code."}
)

// --- 404 Not Found ---
var (
	ErrUserNotFound    = APIError{http.StatusNotFound, "user_not_found", "User doesn't exist."}
	ErrPostNotFound    = APIError{http.StatusNotFound, "post_not_found", "Post not found."}
	ErrBlogNotFound    = APIError{http.StatusNotFound, "blog_not_found", "Blog doesn't exist."}
	ErrTokenNotFound   = APIError{http.StatusNotFound, "token_not_found", "Token not found."}
	ErrMediaNotFound   = APIError{http.StatusNotFound, "media_not_found", "Media not found."}
)

// --- 409 Conflict ---
var (
	ErrUsernameTaken  = APIError{http.StatusConflict, "username_taken", "Username is already taken."}
	ErrSlugTaken      = APIError{http.StatusConflict, "slug_taken", "This slug is already in use."}
	ErrDomainTaken    = APIError{http.StatusConflict, "domain_taken", "This custom domain is already registered."}
)

// --- 410 Gone ---
var (
	ErrPostUnpublished = APIError{http.StatusGone, "post_unpublished", "This post was unpublished by the author."}
	ErrPostRemoved     = APIError{http.StatusGone, "post_removed", "This post has been removed."}
	ErrBlogGone        = APIError{http.StatusGone, "blog_gone", "This blog was unpublished."}
)

// --- 429 Too Many Requests ---
var (
	ErrRateLimited = APIError{http.StatusTooManyRequests, "rate_limited", "Too many requests. Please slow down."}
)

// --- 500 Internal Server Error ---
var (
	ErrInternal = APIError{http.StatusInternalServerError, "server_error", "Something went wrong on our end. Please try again."}
	ErrUnavailable = APIError{http.StatusServiceUnavailable, "unavailable", "Service temporarily unavailable."}
)

// WithMessage returns a copy of the error with a custom message (for dynamic context).
func (e APIError) WithMessage(msg string) APIError {
	e.Message = msg
	return e
}
