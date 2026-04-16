package server

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"

	"github.com/you/inkvault/internal/auth"
	"github.com/you/inkvault/internal/config"
	"github.com/you/inkvault/internal/crypto"
	"github.com/you/inkvault/internal/media"
	"github.com/you/inkvault/internal/middleware"
	"github.com/you/inkvault/internal/posts"
	"github.com/you/inkvault/internal/repository"
	"github.com/you/inkvault/internal/repository/postgres"
	"github.com/you/inkvault/internal/rss"
)

// New wires all dependencies and returns the configured Fiber app.
func New(cfg *config.Config) (*fiber.App, error) {
	// ── Infrastructure ────────────────────────────────────────────────────────
	store, err := postgres.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	rdbOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(rdbOpts)

	encryptor, err := crypto.NewEmailEncryptor(cfg.EmailEncryptionKey)
	if err != nil {
		return nil, err
	}

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := auth.NewService(cfg, store, encryptor)
	authHandler := auth.NewHandler(authSvc)

	postsSvc := posts.NewService(store)
	postsHandler := posts.NewHandler(postsSvc)

	rssSvc := rss.NewService(store, cfg.AllowedOrigins)

	var mediaHandler *media.Handler
	if mediaSvc, err := media.NewService(cfg); err == nil {
		mediaHandler = media.NewHandler(mediaSvc, store)
	}

	// ── App ───────────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:               "InkVault API",
		DisableStartupMessage: false,
		ErrorHandler:          errorHandler,
		ServerHeader:          "",
		BodyLimit:             11 * 1024 * 1024, // 11 MB (media uploads)
	})

	// ── Global middleware ─────────────────────────────────────────────────────
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(middleware.SecurityHeaders())
	app.Use(compress.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Authorization,Accept,X-API-Key",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	// ── Routes ────────────────────────────────────────────────────────────────
	api := app.Group("/api/v1")

	// Health
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "version": "1.0.0"})
	})

	// ── Auth (rate-limited) ───────────────────────────────────────────────────
	authLimiter := middleware.RateLimiter(rdb, cfg.AuthRateLimit, cfg.AuthRateWindow)
	authGroup := api.Group("/auth", authLimiter)
	authGroup.Post("/register",        authHandler.Register)
	authGroup.Post("/login",           authHandler.Login)
	authGroup.Post("/refresh",         authHandler.Refresh)
	authGroup.Post("/logout",          authHandler.Logout)
	authGroup.Post("/forgot-password", authHandler.ForgotPassword)
	authGroup.Post("/reset-password",  authHandler.ResetPassword)
	authGroup.Get("/providers",        func(c *fiber.Ctx) error { return c.JSON(auth.NewOAuthProviders(cfg)) })
	authGroup.Get("/username-check",   func(c *fiber.Ctx) error {
		return c.JSON(authSvc.CheckUsernameAvailability(c.Context(), c.Query("username")))
	})

	// ── Public (optional auth) ────────────────────────────────────────────────
	optAuth := middleware.OptionalAuth(cfg.JWTSecret)
	public := api.Group("/", optAuth)
	public.Get("/blogs/:slug",                 postsHandler.GetBlog)
	public.Get("/blogs/:slug/posts",           postsHandler.ListBlogPosts)
	public.Get("/blogs/:slug/posts/:postSlug", postsHandler.GetPost)
	public.Get("/search",                      postsHandler.SearchPosts)

	// RSS / Atom feeds
	public.Get("/blogs/:slug/feed.xml", func(c *fiber.Ctx) error {
		feed, err := rssSvc.GenerateRSS(c.Context(), c.Params("slug"))
		if err != nil {
			return c.Status(404).SendString("Feed not found")
		}
		c.Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Set("Cache-Control", "public, max-age=3600")
		return c.SendString(feed)
	})
	public.Get("/blogs/:slug/atom.xml", func(c *fiber.Ctx) error {
		feed, err := rssSvc.GenerateAtom(c.Context(), c.Params("slug"))
		if err != nil {
			return c.Status(404).SendString("Feed not found")
		}
		c.Set("Content-Type", "application/atom+xml; charset=utf-8")
		c.Set("Cache-Control", "public, max-age=3600")
		return c.SendString(feed)
	})

	// ── Protected (JWT required) ──────────────────────────────────────────────
	jwtMW := middleware.JWTMiddleware(cfg.JWTSecret)
	protected := api.Group("/", jwtMW)

	protected.Post("/auth/logout-all", authHandler.LogoutAll)
	protected.Get("/me/blogs",         postsHandler.MyBlogs)
	protected.Get("/me/posts",         postsHandler.MyPosts)

	// GDPR export
	protected.Get("/me/export/posts.json", func(c *fiber.Ctx) error {
		userID, _ := c.Locals("userID").(string)
		ps, _ := store.Posts().GetPostsByAuthor(c.Context(), userID, 10000, 0)
		c.Set("Content-Disposition", `attachment; filename="inkvault-posts.json"`)
		return c.JSON(fiber.Map{"posts": ps})
	})

	// Writer+ routes
	writerMW := middleware.RequireRole("writer")
	protected.Post("/blogs",              writerMW, postsHandler.CreateBlog)
	protected.Post("/posts",              writerMW, postsHandler.CreatePost)
	protected.Patch("/posts/:id",         writerMW, postsHandler.UpdatePost)
	protected.Delete("/posts/:id",        writerMW, postsHandler.DeletePost)
	protected.Get("/posts/:id/revisions", postsHandler.GetRevisions)

	// Media
	if mediaHandler != nil {
		protected.Post("/media",     writerMW, mediaHandler.Upload)
		protected.Delete("/media/*", writerMW, mediaHandler.Delete)
		protected.Get("/me/media",   func(c *fiber.Ctx) error {
			uploaderID, _ := c.Locals("userID").(string)
			items, _ := store.Media().ListMediaByUploader(c.Context(), uploaderID, 50, 0)
			return c.JSON(fiber.Map{"media": items})
		})
	}

	// Admin routes
	adminMW := middleware.RequireRole("admin")
	admin := protected.Group("/admin", adminMW)
	admin.Get("/users", func(c *fiber.Ctx) error {
		limit := c.QueryInt("limit", 20)
		offset := c.QueryInt("offset", 0)
		u, err := store.Users().ListUsers(c.Context(), limit, offset)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "server_error"})
		}
		pub := make([]interface{}, len(u))
		for i, user := range u {
			pub[i] = user.ToPublic()
		}
		return c.JSON(fiber.Map{"users": pub, "total": len(pub)})
	})
	admin.Get("/audit", func(c *fiber.Ctx) error {
		logs, err := store.Audit().ListLogs(c.Context(), repository.AuditFilters{
			Limit: c.QueryInt("limit", 50), Offset: c.QueryInt("offset", 0),
		})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "server_error"})
		}
		return c.JSON(fiber.Map{"logs": logs})
	})

	return app, nil
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{"error": "error", "message": err.Error()})
}
