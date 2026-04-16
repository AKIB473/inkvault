// Package repository defines the data access interface.
// Inspired by WriteFreely's writestore interface — the entire DB is abstracted
// behind this interface, making it easy to swap implementations or mock in tests.
package repository

import (
	"context"

	"github.com/you/inkvault/internal/domain"
)

// UserRepository handles all user data operations.
type UserRepository interface {
	// Auth & account
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserByEmailHash(ctx context.Context, emailHash string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	UpdateEncryptedEmail(ctx context.Context, userID, encryptedEmail string) error
	UpdatePasswordHash(ctx context.Context, userID, hash string) error
	UpdateRole(ctx context.Context, userID string, role domain.Role) error
	UpdateStatus(ctx context.Context, userID string, status domain.UserStatus) error
	Enable2FA(ctx context.Context, userID string) error
	Disable2FA(ctx context.Context, userID string) error
	// Hard delete — GDPR compliant, cascades all user data
	DeleteAccount(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error)
}

// TokenRepository handles single-use tokens.
type TokenRepository interface {
	CreateToken(ctx context.Context, token *domain.Token) error
	GetTokenByHash(ctx context.Context, hash string, tokenType domain.TokenType) (*domain.Token, error)
	MarkTokenUsed(ctx context.Context, tokenID string) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteTokensByUser(ctx context.Context, userID string, tokenType domain.TokenType) error
}

// SessionRepository handles refresh token sessions.
type SessionRepository interface {
	CreateSession(ctx context.Context, session *domain.Session) error
	GetSessionByRefreshHash(ctx context.Context, hash string) (*domain.Session, error)
	UpdateSessionLastSeen(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteAllUserSessions(ctx context.Context, userID string) error
	ListUserSessions(ctx context.Context, userID string) ([]*domain.Session, error)
}

// BlogRepository handles blog (collection) operations.
type BlogRepository interface {
	CreateBlog(ctx context.Context, blog *domain.Blog) error
	GetBlogByID(ctx context.Context, id string) (*domain.Blog, error)
	GetBlogBySlug(ctx context.Context, slug string) (*domain.Blog, error)
	GetBlogByDomain(ctx context.Context, domain string) (*domain.Blog, error)
	GetBlogsByOwner(ctx context.Context, ownerID string) ([]*domain.Blog, error)
	UpdateBlog(ctx context.Context, blog *domain.Blog) error
	DeleteBlog(ctx context.Context, id, ownerID string) error
}

// PostRepository handles post CRUD and revisions.
type PostRepository interface {
	CreatePost(ctx context.Context, post *domain.Post) error
	GetPostByID(ctx context.Context, id string) (*domain.Post, error)
	GetPostBySlug(ctx context.Context, blogID, slug string) (*domain.Post, error)
	GetPostsByBlog(ctx context.Context, blogID string, status domain.PostStatus, limit, offset int) ([]*domain.Post, error)
	GetPostsByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*domain.Post, error)
	UpdatePost(ctx context.Context, post *domain.Post) error
	DeletePost(ctx context.Context, id, authorID string) error
	IncrementViews(ctx context.Context, postID string) error
	SearchPosts(ctx context.Context, query string, limit, offset int) ([]*domain.Post, error)

	// Revisions
	CreateRevision(ctx context.Context, rev *domain.PostRevision) error
	GetRevisionsByPost(ctx context.Context, postID string) ([]*domain.PostRevision, error)
	GetRevisionByID(ctx context.Context, id string) (*domain.PostRevision, error)
}

// AuditRepository stores audit log entries.
type AuditRepository interface {
	CreateLog(ctx context.Context, entry *domain.AuditLog) error
	ListLogs(ctx context.Context, filters AuditFilters) ([]*domain.AuditLog, error)
}

// AuditFilters for querying audit logs.
type AuditFilters struct {
	ActorID      string
	ResourceType string
	Action       string
	Limit        int
	Offset       int
}

// Store bundles all repositories — passed through the app via dependency injection.
type Store interface {
	Users() UserRepository
	Tokens() TokenRepository
	Sessions() SessionRepository
	Blogs() BlogRepository
	Posts() PostRepository
	Audit() AuditRepository
	Media() MediaRepository
}
