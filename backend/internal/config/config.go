package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Env  string
	Port string

	DatabaseURL string

	RedisURL string

	JWTSecret      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration

	// AES-256-GCM key for email encryption (32 bytes, hex-encoded in env)
	EmailEncryptionKey string

	ResendAPIKey string
	EmailFrom    string

	S3Endpoint   string
	S3Region     string
	S3Bucket     string
	S3AccessKey  string
	S3SecretKey  string
	S3PublicURL  string

	OpenRegistration bool
	InviteOnly       bool

	AuthRateLimit  int
	AuthRateWindow time.Duration

	ActivityPubEnabled bool
	AllowedOrigins     string

	// OAuth providers (all opt-in via config — WriteFreely OAuthButtons pattern)
	GitHubOAuthClientID     string
	GitHubOAuthClientSecret string
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
	GenericOAuthClientID     string
	GenericOAuthClientSecret string
	GenericOAuthName         string
	GenericOAuthAuthURL      string
	GenericOAuthTokenURL     string
}

func Load() (*Config, error) {
	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	authWindow, err := time.ParseDuration(getEnv("AUTH_RATE_WINDOW", "1h"))
	if err != nil {
		return nil, fmt.Errorf("invalid AUTH_RATE_WINDOW: %w", err)
	}

	cfg := &Config{
		Env:  getEnv("APP_ENV", "development"),
		Port: getEnv("APP_PORT", "8080"),

		DatabaseURL: requireEnv("DATABASE_URL"),

		RedisURL: requireEnv("REDIS_URL"),

		JWTSecret:     requireEnv("JWT_SECRET"),
		JWTAccessTTL:  accessTTL,
		JWTRefreshTTL: refreshTTL,

		EmailEncryptionKey: requireEnv("EMAIL_ENCRYPTION_KEY"),

		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		EmailFrom:    getEnv("EMAIL_FROM", "noreply@localhost"),

		S3Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3Region:    getEnv("S3_REGION", "us-east-1"),
		S3Bucket:    getEnv("S3_BUCKET", "inkvault-media"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey: getEnv("S3_SECRET_KEY", ""),
		S3PublicURL: getEnv("S3_PUBLIC_URL", ""),

		OpenRegistration: getEnvBool("OPEN_REGISTRATION", true),
		InviteOnly:       getEnvBool("INVITE_ONLY", false),

		AuthRateLimit:  getEnvInt("AUTH_RATE_LIMIT", 5),
		AuthRateWindow: authWindow,

		ActivityPubEnabled: getEnvBool("ACTIVITYPUB_ENABLED", false),
		AllowedOrigins:     getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),

		GitHubOAuthClientID:      getEnv("GITHUB_OAUTH_CLIENT_ID", ""),
		GitHubOAuthClientSecret:  getEnv("GITHUB_OAUTH_CLIENT_SECRET", ""),
		GoogleOAuthClientID:      getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleOAuthClientSecret:  getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		GenericOAuthClientID:     getEnv("GENERIC_OAUTH_CLIENT_ID", ""),
		GenericOAuthClientSecret: getEnv("GENERIC_OAUTH_CLIENT_SECRET", ""),
		GenericOAuthName:         getEnv("GENERIC_OAUTH_NAME", "SSO"),
		GenericOAuthAuthURL:      getEnv("GENERIC_OAUTH_AUTH_URL", ""),
		GenericOAuthTokenURL:     getEnv("GENERIC_OAUTH_TOKEN_URL", ""),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	switch v {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var i int
	if _, err := fmt.Sscanf(v, "%d", &i); err != nil {
		return fallback
	}
	return i
}
