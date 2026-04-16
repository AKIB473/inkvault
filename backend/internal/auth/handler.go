package auth

import (
	"github.com/gofiber/fiber/v2"
)

// Handler exposes auth endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register godoc
// POST /api/v1/auth/register
func (h *Handler) Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_body",
			"message": "Could not parse request body",
		})
	}

	user, accessToken, err := h.svc.Register(c.Context(), input)
	if err != nil {
		switch err {
		case ErrHoneypotTriggered:
			// Return 200 to confuse bots
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Check your email to verify your account"})
		case ErrUsernameExists:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username_taken", "message": "Username already taken"})
		case ErrRegistrationClosed:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "registration_closed", "message": "Registration is currently closed"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error", "message": "Registration failed"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":         user.ToPublic(),
		"access_token": accessToken,
		"message":      "Account created successfully",
	})
}

// Login godoc
// POST /api/v1/auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_body",
			"message": "Could not parse request body",
		})
	}

	result, err := h.svc.Login(c.Context(), input, c.IP(), c.Get("User-Agent"))
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_credentials", "message": "Invalid username or password"})
		case ErrUserBanned:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "account_banned", "message": "Your account has been suspended"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error", "message": "Login failed"})
		}
	}

	if result.Requires2FA {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"requires_2fa": true,
			"message":      "Enter your two-factor authentication code",
		})
	}

	// Set refresh token as HttpOnly cookie (more secure than localStorage)
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   int(h.svc.cfg.JWTRefreshTTL.Seconds()),
		Path:     "/api/v1/auth",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":         result.User.ToPublic(),
		"access_token": result.AccessToken,
	})
}

// Refresh godoc
// POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		// Also accept from body (mobile clients)
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.BodyParser(&body)
		refreshToken = body.RefreshToken
	}

	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized", "message": "Refresh token required"})
	}

	result, err := h.svc.RefreshTokens(c.Context(), refreshToken, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized", "message": "Invalid or expired refresh token"})
	}

	// Rotate cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   int(h.svc.cfg.JWTRefreshTTL.Seconds()),
		Path:     "/api/v1/auth",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": result.AccessToken,
	})
}

// Logout godoc
// POST /api/v1/auth/logout
func (h *Handler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		_ = h.svc.Logout(c.Context(), refreshToken)
	}

	// Clear cookie
	c.ClearCookie("refresh_token")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Logged out successfully"})
}

// ForgotPassword godoc
// POST /api/v1/auth/forgot-password
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid_body"})
	}

	// Always return 200 regardless of whether user exists
	// Ghost pattern: don't reveal user existence via reset endpoint
	_, _ = h.svc.GeneratePasswordResetToken(c.Context(), body.Username)
	// TODO: send email with reset link

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If that account exists, a reset link has been sent to the registered email address.",
	})
}

// ResetPassword godoc
// POST /api/v1/auth/reset-password
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var input ResetPasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid_body"})
	}

	if err := h.svc.ResetPassword(c.Context(), input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "reset_failed",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password reset successfully. Please log in with your new password.",
	})
}

// LogoutAll godoc — revoke all sessions (all devices)
// POST /api/v1/auth/logout-all
func (h *Handler) LogoutAll(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	if err := h.svc.LogoutAll(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
	}
	c.ClearCookie("refresh_token")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "All sessions revoked"})
}
