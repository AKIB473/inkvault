// Package email sends transactional email via Resend API.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Service struct {
	apiKey  string
	from    string
	appName string
	appURL  string
	client  *http.Client
}

func NewService(apiKey, from, appName, appURL string) *Service {
	return &Service{
		apiKey:  apiKey,
		from:    from,
		appName: appName,
		appURL:  appURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type sendRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

// SendPasswordReset sends a password reset email with a single-use link.
// Ghost pattern: link is valid 24h, shows expiry in email body.
func (s *Service) SendPasswordReset(ctx context.Context, to, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.appURL, resetToken)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:560px;margin:0 auto;padding:20px;color:#111">
  <h2>🖋️ Reset your %s password</h2>
  <p>Someone (hopefully you) requested a password reset for your account.</p>
  <p style="margin:24px 0">
    <a href="%s" style="display:inline-block;padding:12px 24px;background:#2563eb;color:white;border-radius:8px;text-decoration:none;font-weight:600">
      Reset password
    </a>
  </p>
  <p style="color:#666;font-size:14px">This link expires in <strong>24 hours</strong> and can only be used once.</p>
  <p style="color:#666;font-size:14px">If you didn't request this, you can safely ignore this email.</p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#999;font-size:12px">%s · Your email is encrypted and never shared.</p>
</body>
</html>`, s.appName, resetURL, s.appName)

	text := fmt.Sprintf("Reset your %s password:\n%s\n\nThis link expires in 24 hours.", s.appName, resetURL)

	return s.send(ctx, to, fmt.Sprintf("Reset your %s password", s.appName), html, text)
}

// SendOTP sends a one-time password for 2FA / device verification.
// Ghost pattern: clear subject with code, expires in 10 minutes.
func (s *Service) SendOTP(ctx context.Context, to, code, username string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:560px;margin:0 auto;padding:20px;color:#111">
  <h2>🖋️ Your %s verification code</h2>
  <p>Hi %s, use this code to complete your sign-in:</p>
  <div style="margin:24px 0;padding:20px;background:#f3f4f6;border-radius:8px;text-align:center">
    <span style="font-size:36px;font-weight:700;letter-spacing:8px;color:#1e293b">%s</span>
  </div>
  <p style="color:#666;font-size:14px">This code expires in <strong>10 minutes</strong>.</p>
  <p style="color:#666;font-size:14px">If you didn't request this, your password may be compromised. Change it immediately.</p>
</body>
</html>`, s.appName, username, code)

	text := fmt.Sprintf("Your %s verification code: %s\n\nExpires in 10 minutes.", s.appName, code)

	return s.send(ctx, to, fmt.Sprintf("%s is your %s sign-in code", code, s.appName), html, text)
}

// SendInvite sends an invite email with a signup link.
// Ghost pattern: show inviter name + blog name + single-use link.
func (s *Service) SendInvite(ctx context.Context, to, inviteToken, inviterName, blogName string) error {
	signupURL := fmt.Sprintf("%s/register?invite=%s", s.appURL, inviteToken)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:560px;margin:0 auto;padding:20px;color:#111">
  <h2>🖋️ You've been invited to %s</h2>
  <p><strong>%s</strong> has invited you to join <strong>%s</strong> on %s.</p>
  <p style="margin:24px 0">
    <a href="%s" style="display:inline-block;padding:12px 24px;background:#2563eb;color:white;border-radius:8px;text-decoration:none;font-weight:600">
      Accept invitation
    </a>
  </p>
  <p style="color:#666;font-size:14px">This invitation expires in <strong>7 days</strong>.</p>
</body>
</html>`, blogName, inviterName, blogName, s.appName, signupURL)

	text := fmt.Sprintf("%s invited you to %s on %s:\n%s", inviterName, blogName, s.appName, signupURL)

	return s.send(ctx, to, fmt.Sprintf("%s has invited you to %s", inviterName, blogName), html, text)
}

// SendWelcome sends a welcome email after registration.
func (s *Service) SendWelcome(ctx context.Context, to, username string) error {
	dashURL := fmt.Sprintf("%s/dashboard", s.appURL)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:560px;margin:0 auto;padding:20px;color:#111">
  <h2>Welcome to %s, %s! 🖋️</h2>
  <p>Your account is ready. Start writing in seconds.</p>
  <p style="margin:24px 0">
    <a href="%s" style="display:inline-block;padding:12px 24px;background:#2563eb;color:white;border-radius:8px;text-decoration:none;font-weight:600">
      Go to dashboard
    </a>
  </p>
  <p style="color:#666;font-size:14px">Your email is encrypted at rest. We'll never share it.</p>
</body>
</html>`, s.appName, username, dashURL)

	text := fmt.Sprintf("Welcome to %s, %s!\n%s", s.appName, username, dashURL)

	return s.send(ctx, to, fmt.Sprintf("Welcome to %s", s.appName), html, text)
}

func (s *Service) send(ctx context.Context, to, subject, html, text string) error {
	if s.apiKey == "" {
		// Dev mode: log instead of send
		fmt.Printf("[EMAIL DEV] To: %s | Subject: %s\n", to, subject)
		return nil
	}

	body, _ := json.Marshal(sendRequest{
		From: s.from, To: to,
		Subject: subject, HTML: html, Text: text,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	return nil
}
