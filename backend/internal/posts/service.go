// Package posts handles all post and blog business logic.
package posts

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/you/inkvault/internal/domain"
	"github.com/you/inkvault/internal/repository"
)

type Service struct {
	store repository.Store
}

func NewService(store repository.Store) *Service {
	return &Service{store: store}
}

// --- Blog operations ---

type CreateBlogInput struct {
	Title       string `json:"title"       validate:"required,min=1,max=255"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Visibility  string `json:"visibility"`
}

func (s *Service) CreateBlog(ctx context.Context, input CreateBlogInput, ownerID string) (*domain.Blog, error) {
	slug := slugify(input.Title)
	if slug == "" {
		return nil, errors.New("invalid blog title — could not generate slug")
	}

	lang := input.Language
	if lang == "" {
		lang = "en"
	}

	vis := parseVisibility(input.Visibility)

	blog := &domain.Blog{
		ID:          uuid.New().String(),
		OwnerID:     ownerID,
		Slug:        slug,
		Title:       input.Title,
		Description: input.Description,
		Visibility:  vis,
		Language:    lang,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.store.Blogs().CreateBlog(ctx, blog); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			// Append UUID suffix for uniqueness
			blog.Slug = slug + "-" + uuid.New().String()[:8]
			if err2 := s.store.Blogs().CreateBlog(ctx, blog); err2 != nil {
				return nil, err2
			}
		} else {
			return nil, err
		}
	}

	return blog, nil
}

func (s *Service) GetBlog(ctx context.Context, slugOrID string) (*domain.Blog, error) {
	blog, err := s.store.Blogs().GetBlogBySlug(ctx, slugOrID)
	if err != nil {
		blog, err = s.store.Blogs().GetBlogByID(ctx, slugOrID)
	}
	return blog, err
}

func (s *Service) GetUserBlogs(ctx context.Context, ownerID string) ([]*domain.Blog, error) {
	return s.store.Blogs().GetBlogsByOwner(ctx, ownerID)
}

// --- Post operations ---

type CreatePostInput struct {
	BlogID      string   `json:"blog_id"      validate:"required"`
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	ContentType string   `json:"content_type"` // "tiptap" | "markdown"
	Excerpt     string   `json:"excerpt"`
	Tags        []string `json:"tags"`
	Visibility  string   `json:"visibility"`
	Status      string   `json:"status"` // "draft" | "published"
	CoverImage  string   `json:"cover_image"`
	SEOTitle    string   `json:"seo_title"`
	SEODesc     string   `json:"seo_description"`
	OGImage     string   `json:"og_image"`
	CanonicalURL string  `json:"canonical_url"`
}

func (s *Service) CreatePost(ctx context.Context, input CreatePostInput, authorID string) (*domain.Post, error) {
	// Verify author has access to the blog
	blog, err := s.store.Blogs().GetBlogByID(ctx, input.BlogID)
	if err != nil {
		return nil, errors.New("blog not found")
	}
	if blog.OwnerID != authorID {
		// TODO: check if author is a contributor to this blog
		return nil, errors.New("you don't have permission to post to this blog")
	}

	slug := slugify(input.Title)
	if slug == "" {
		slug = uuid.New().String()[:8]
	}

	contentType := input.ContentType
	if contentType == "" {
		contentType = "tiptap"
	}

	status := domain.PostStatusDraft
	if input.Status == "published" {
		status = domain.PostStatusPublished
	}

	vis := parseVisibility(input.Visibility)

	now := time.Now()
	post := &domain.Post{
		ID:          uuid.New().String(),
		BlogID:      input.BlogID,
		AuthorID:    authorID,
		Title:       input.Title,
		Slug:        slug,
		Excerpt:     input.Excerpt,
		Content:     input.Content,
		ContentType: contentType,
		Status:      status,
		Visibility:  vis,
		Tags:        sanitizeTags(input.Tags),
		CoverImage:  input.CoverImage,
		SEOTitle:    input.SEOTitle,
		SEODescription: input.SEODesc,
		OGImage:     input.OGImage,
		CanonicalURL: input.CanonicalURL,
		ReadingTimeMins: estimateReadingTime(input.Content),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if status == domain.PostStatusPublished {
		post.PublishedAt = &now
	}

	if err := s.store.Posts().CreatePost(ctx, post); err != nil {
		return nil, err
	}

	// Create initial revision
	rev := &domain.PostRevision{
		ID:        uuid.New().String(),
		PostID:    post.ID,
		Content:   post.Content,
		EditedBy:  authorID,
		CreatedAt: now,
	}
	_ = s.store.Posts().CreateRevision(ctx, rev) // Non-fatal

	return post, nil
}

type UpdatePostInput struct {
	Title       *string  `json:"title"`
	Content     *string  `json:"content"`
	Excerpt     *string  `json:"excerpt"`
	Tags        []string `json:"tags"`
	Visibility  *string  `json:"visibility"`
	Status      *string  `json:"status"`
	CoverImage  *string  `json:"cover_image"`
	SEOTitle    *string  `json:"seo_title"`
	SEODesc     *string  `json:"seo_description"`
	OGImage     *string  `json:"og_image"`
}

func (s *Service) UpdatePost(ctx context.Context, postID string, input UpdatePostInput, editorID string) (*domain.Post, error) {
	post, err := s.store.Posts().GetPostByID(ctx, postID)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Permission: author or editor/admin (simplified — full check in middleware)
	if post.AuthorID != editorID {
		editor, err := s.store.Users().GetUserByID(ctx, editorID)
		if err != nil || !editor.CanEdit() {
			return nil, errors.New("you don't have permission to edit this post")
		}
	}

	// Save revision before updating
	rev := &domain.PostRevision{
		ID:        uuid.New().String(),
		PostID:    post.ID,
		Content:   post.Content,
		EditedBy:  editorID,
		CreatedAt: time.Now(),
	}
	_ = s.store.Posts().CreateRevision(ctx, rev)

	// Apply updates
	if input.Title != nil {
		post.Title = *input.Title
		post.Slug = slugify(*input.Title)
	}
	if input.Content != nil {
		post.Content = *input.Content
		post.ReadingTimeMins = estimateReadingTime(*input.Content)
	}
	if input.Excerpt != nil {
		post.Excerpt = *input.Excerpt
	}
	if len(input.Tags) > 0 {
		post.Tags = sanitizeTags(input.Tags)
	}
	if input.Visibility != nil {
		post.Visibility = parseVisibility(*input.Visibility)
	}
	if input.Status != nil {
		newStatus := domain.PostStatus(*input.Status)
		if post.Status != domain.PostStatusPublished && newStatus == domain.PostStatusPublished {
			now := time.Now()
			post.PublishedAt = &now
		}
		post.Status = newStatus
	}
	if input.CoverImage != nil {
		post.CoverImage = *input.CoverImage
	}
	if input.SEOTitle != nil {
		post.SEOTitle = *input.SEOTitle
	}
	if input.SEODesc != nil {
		post.SEODescription = *input.SEODesc
	}
	if input.OGImage != nil {
		post.OGImage = *input.OGImage
	}
	post.UpdatedAt = time.Now()

	if err := s.store.Posts().UpdatePost(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) DeletePost(ctx context.Context, postID, authorID string) error {
	return s.store.Posts().DeletePost(ctx, postID, authorID)
}

func (s *Service) GetPost(ctx context.Context, blogID, slugOrID string, viewerID string) (*domain.Post, error) {
	post, err := s.store.Posts().GetPostBySlug(ctx, blogID, slugOrID)
	if err != nil {
		post, err = s.store.Posts().GetPostByID(ctx, slugOrID)
		if err != nil {
			return nil, err
		}
	}

	// Visibility check
	switch post.Visibility {
	case domain.VisibilityPublic, domain.VisibilityUnlisted:
		// Anyone can read
	case domain.VisibilityMembers:
		if viewerID == "" {
			return nil, errors.New("this post is for members only")
		}
	case domain.VisibilityPrivate:
		if viewerID != post.AuthorID {
			return nil, errors.New("this post is private")
		}
	}

	// Increment view count async (non-blocking)
	go func() {
		_ = s.store.Posts().IncrementViews(context.Background(), post.ID)
	}()

	return post, nil
}

func (s *Service) ListBlogPosts(ctx context.Context, blogID string, publishedOnly bool, limit, offset int) ([]*domain.Post, error) {
	status := domain.PostStatus("")
	if publishedOnly {
		status = domain.PostStatusPublished
	}
	return s.store.Posts().GetPostsByBlog(ctx, blogID, status, limit, offset)
}

func (s *Service) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*domain.Post, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query cannot be empty")
	}
	return s.store.Posts().SearchPosts(ctx, query, limit, offset)
}

// --- Helpers ---

// slugify converts a string to a URL-safe slug.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash && b.Len() > 0 {
			b.WriteRune('-')
			prevDash = true
		}
	}
	result := strings.TrimRight(b.String(), "-")
	if len(result) > 200 {
		result = result[:200]
	}
	return result
}

// estimateReadingTime returns estimated reading time in minutes.
// Average reading speed: 200 words/min.
func estimateReadingTime(content string) int {
	words := len(strings.Fields(content))
	mins := words / 200
	if mins < 1 {
		return 1
	}
	return mins
}

// sanitizeTags trims and lowercases tags, removes duplicates.
func sanitizeTags(tags []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" && !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}
	return result
}

func parseVisibility(s string) domain.PostVisibility {
	switch s {
	case "unlisted":
		return domain.VisibilityUnlisted
	case "members_only":
		return domain.VisibilityMembers
	case "private":
		return domain.VisibilityPrivate
	case "password_protected":
		return domain.VisibilityPassword
	default:
		return domain.VisibilityPublic
	}
}
