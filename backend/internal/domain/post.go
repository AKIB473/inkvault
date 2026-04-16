package domain

import "time"

// PostVisibility controls who can see a post (learned from WriteFreely's bitmask).
type PostVisibility int

const (
	VisibilityPublic    PostVisibility = 0 // Anyone can read
	VisibilityUnlisted  PostVisibility = 1 // Anyone with link, not indexed
	VisibilityMembers   PostVisibility = 2 // Logged-in members only
	VisibilityPrivate   PostVisibility = 3 // Author only
	VisibilityPassword  PostVisibility = 4 // Password protected
)

// PostStatus represents the editorial state.
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusScheduled PostStatus = "scheduled"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
)

// Post is the core article/post entity.
type Post struct {
	ID           string         `json:"id"`
	BlogID       string         `json:"blog_id"`
	AuthorID     string         `json:"author_id"`
	Title        string         `json:"title"`
	Slug         string         `json:"slug"`
	Excerpt      string         `json:"excerpt"`
	Content      string         `json:"content"`       // Rich text JSON (Tiptap) or Markdown
	ContentType  string         `json:"content_type"`  // "tiptap" | "markdown"
	Status       PostStatus     `json:"status"`
	Visibility   PostVisibility `json:"visibility"`
	PasswordHash string         `json:"-"`             // For password-protected posts
	IsPinned     bool           `json:"is_pinned"`
	Tags         []string       `json:"tags"`
	CoverImage   string         `json:"cover_image,omitempty"`

	// SEO (learned from Comma)
	SEOTitle       string `json:"seo_title,omitempty"`
	SEODescription string `json:"seo_description,omitempty"`
	OGImage        string `json:"og_image,omitempty"`
	CanonicalURL   string `json:"canonical_url,omitempty"`

	// Metadata
	Views          int64     `json:"views"`
	ReadingTimeMins int      `json:"reading_time_mins"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Newsletter (learned from Comma)
	LastNewsletterSentAt *time.Time `json:"last_newsletter_sent_at,omitempty"`

	// Populated on read (not stored)
	Author *PublicUser `json:"author,omitempty"`
}

// PostRevision stores a snapshot of a post at a given time (learned from Ghost + goblog).
type PostRevision struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	Content   string    `json:"content"`
	EditedBy  string    `json:"edited_by"`
	CreatedAt time.Time `json:"created_at"`
}

// Blog (Collection in WriteFreely) — each user can have multiple blogs.
type Blog struct {
	ID          string         `json:"id"`
	OwnerID     string         `json:"owner_id"`
	Slug        string         `json:"slug"`        // URL identifier
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Domain      string         `json:"domain,omitempty"` // Custom domain
	Visibility  PostVisibility `json:"visibility"`
	PasswordHash string        `json:"-"`
	StyleSheet  string         `json:"style_sheet,omitempty"` // Custom CSS
	Language    string         `json:"language"`
	RTL         bool           `json:"rtl"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// Populated on read
	TotalPosts int         `json:"total_posts,omitempty"`
	Owner      *PublicUser `json:"owner,omitempty"`
}
