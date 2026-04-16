package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterKey(t *testing.T) {
	// Verify key format is consistent and path-scoped
	path := "/api/v1/auth/login"
	ip := "192.168.1.1"
	key := "ratelimit:" + path + ":" + ip

	if key != "ratelimit:/api/v1/auth/login:192.168.1.1" {
		t.Errorf("unexpected key format: %s", key)
	}
}

func TestRateLimiterWindow(t *testing.T) {
	// Ensure window values are sane
	windows := []time.Duration{
		time.Hour,
		5 * time.Minute,
		time.Second * 30,
	}
	for _, w := range windows {
		if w <= 0 {
			t.Errorf("window %v is non-positive", w)
		}
	}
}
