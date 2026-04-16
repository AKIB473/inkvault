package posts

import (
	"github.com/gofiber/fiber/v2"
	"github.com/you/inkvault/internal/apierr"
	"github.com/you/inkvault/internal/repository"
)

// Handler exposes post and blog HTTP endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ── Blog endpoints ────────────────────────────────────────────────────────────

// CreateBlog godoc — POST /api/v1/blogs
func (h *Handler) CreateBlog(c *fiber.Ctx) error {
	ownerID, _ := c.Locals("userID").(string)

	var input CreateBlogInput
	if err := c.BodyParser(&input); err != nil {
		return apierr.ErrBadJSON.FiberResponse(c)
	}

	blog, err := h.svc.CreateBlog(c.Context(), input, ownerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "create_failed", "message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"blog": blog})
}

// GetBlog godoc — GET /api/v1/blogs/:slug
func (h *Handler) GetBlog(c *fiber.Ctx) error {
	slug := c.Params("slug")
	blog, err := h.svc.GetBlog(c.Context(), slug)
	if err != nil {
		return apierr.ErrBlogNotFound.FiberResponse(c)
	}
	return c.JSON(fiber.Map{"blog": blog})
}

// MyBlogs godoc — GET /api/v1/me/blogs
func (h *Handler) MyBlogs(c *fiber.Ctx) error {
	ownerID, _ := c.Locals("userID").(string)
	blogs, err := h.svc.GetUserBlogs(c.Context(), ownerID)
	if err != nil {
		return apierr.ErrInternal.FiberResponse(c)
	}
	return c.JSON(fiber.Map{"blogs": blogs, "total": len(blogs)})
}

// ── Post endpoints ────────────────────────────────────────────────────────────

// CreatePost godoc — POST /api/v1/posts
func (h *Handler) CreatePost(c *fiber.Ctx) error {
	authorID, _ := c.Locals("userID").(string)

	var input CreatePostInput
	if err := c.BodyParser(&input); err != nil {
		return apierr.ErrBadJSON.FiberResponse(c)
	}

	post, err := h.svc.CreatePost(c.Context(), input, authorID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "create_failed", "message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"post": post})
}

// GetPost godoc — GET /api/v1/blogs/:slug/posts/:postSlug
func (h *Handler) GetPost(c *fiber.Ctx) error {
	blogSlug := c.Params("slug")
	postSlug := c.Params("postSlug")
	viewerID, _ := c.Locals("userID").(string) // Optional auth

	// Resolve blog
	blog, err := h.svc.GetBlog(c.Context(), blogSlug)
	if err != nil {
		return apierr.ErrBlogNotFound.FiberResponse(c)
	}

	post, err := h.svc.GetPost(c.Context(), blog.ID, postSlug, viewerID)
	if err != nil {
		switch err.Error() {
		case "this post is for members only":
			return apierr.ErrForbidden.WithMessage(err.Error()).FiberResponse(c)
		case "this post is private":
			return apierr.ErrForbidden.WithMessage(err.Error()).FiberResponse(c)
		default:
			return apierr.ErrPostNotFound.FiberResponse(c)
		}
	}

	return c.JSON(fiber.Map{"post": post})
}

// ListBlogPosts godoc — GET /api/v1/blogs/:slug/posts
func (h *Handler) ListBlogPosts(c *fiber.Ctx) error {
	blogSlug := c.Params("slug")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	if limit > 100 {
		limit = 100
	}

	blog, err := h.svc.GetBlog(c.Context(), blogSlug)
	if err != nil {
		return apierr.ErrBlogNotFound.FiberResponse(c)
	}

	posts, err := h.svc.ListBlogPosts(c.Context(), blog.ID, true, limit, offset)
	if err != nil {
		return apierr.ErrInternal.FiberResponse(c)
	}

	return c.JSON(fiber.Map{
		"posts":  posts,
		"total":  len(posts),
		"limit":  limit,
		"offset": offset,
	})
}

// MyPosts godoc — GET /api/v1/me/posts
func (h *Handler) MyPosts(c *fiber.Ctx) error {
	authorID, _ := c.Locals("userID").(string)
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	if limit > 100 {
		limit = 100
	}

	posts, err := h.svc.store.Posts().GetPostsByAuthor(c.Context(), authorID, limit, offset)
	if err != nil {
		return apierr.ErrInternal.FiberResponse(c)
	}

	return c.JSON(fiber.Map{"posts": posts, "total": len(posts)})
}

// UpdatePost godoc — PATCH /api/v1/posts/:id
func (h *Handler) UpdatePost(c *fiber.Ctx) error {
	postID := c.Params("id")
	editorID, _ := c.Locals("userID").(string)

	var input UpdatePostInput
	if err := c.BodyParser(&input); err != nil {
		return apierr.ErrBadJSON.FiberResponse(c)
	}

	post, err := h.svc.UpdatePost(c.Context(), postID, input, editorID)
	if err != nil {
		if err.Error() == "post not found" {
			return apierr.ErrPostNotFound.FiberResponse(c)
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "update_failed", "message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"post": post})
}

// DeletePost godoc — DELETE /api/v1/posts/:id
func (h *Handler) DeletePost(c *fiber.Ctx) error {
	postID := c.Params("id")
	authorID, _ := c.Locals("userID").(string)

	if err := h.svc.DeletePost(c.Context(), postID, authorID); err != nil {
		if err == repository.ErrNotFound {
			return apierr.ErrPostNotFound.FiberResponse(c)
		}
		return apierr.ErrForbiddenEdit.FiberResponse(c)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// SearchPosts godoc — GET /api/v1/search?q=...
func (h *Handler) SearchPosts(c *fiber.Ctx) error {
	q := c.Query("q")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	posts, err := h.svc.SearchPosts(c.Context(), q, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "search_error", "message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"posts": posts, "query": q})
}

// GetRevisions godoc — GET /api/v1/posts/:id/revisions
func (h *Handler) GetRevisions(c *fiber.Ctx) error {
	postID := c.Params("id")
	revisions, err := h.svc.store.Posts().GetRevisionsByPost(c.Context(), postID)
	if err != nil {
		return apierr.ErrInternal.FiberResponse(c)
	}
	return c.JSON(fiber.Map{"revisions": revisions})
}
