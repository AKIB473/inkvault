// Seed script — creates dev data: owner user + sample blog + 2 posts.
package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Check if already seeded
	var count int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if count > 0 {
		fmt.Println("Already seeded. Skipping.")
		return
	}

	// ── Owner user ───────────────────────────────────────────────────────────
	ownerID := uuid.New().String()
	hash, _ := bcrypt.GenerateFromPassword([]byte("devpassword123"), 12)
	// Fake encrypted email (hex of "admin@localhost")
	fakeEmail := hex.EncodeToString([]byte("admin@localhost"))

	_, err = pool.Exec(ctx, `
		INSERT INTO users (id,username,display_name,email_encrypted,password_hash,role,status,two_fa_enabled,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,'owner',0,false,$6,$6)`,
		ownerID, "admin", "Admin", fakeEmail, string(hash), time.Now(),
	)
	if err != nil {
		log.Fatalf("seed owner: %v", err)
	}
	fmt.Printf("✅ Owner user created: admin / devpassword123\n")

	// ── Sample blog ──────────────────────────────────────────────────────────
	blogID := uuid.New().String()
	_, err = pool.Exec(ctx, `
		INSERT INTO blogs (id,owner_id,slug,title,description,visibility,language,rtl,created_at,updated_at)
		VALUES ($1,$2,'my-blog','My Blog','A sample blog for development.','public','en',false,$3,$3)`,
		blogID, ownerID, time.Now(),
	)
	if err != nil {
		log.Fatalf("seed blog: %v", err)
	}
	fmt.Printf("✅ Blog created: /my-blog\n")

	// ── Sample posts ─────────────────────────────────────────────────────────
	now := time.Now()
	posts := []struct {
		title   string
		slug    string
		excerpt string
		content string
		tags    []string
	}{
		{
			title:   "Welcome to InkVault",
			slug:    "welcome-to-inkvault",
			excerpt: "InkVault is a secure, privacy-first blogging platform built with Go and Next.js.",
			content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Welcome to InkVault"}]},{"type":"paragraph","content":[{"type":"text","text":"This is your first post. InkVault is built with privacy in mind — your email is encrypted at rest, there are no trackers, and you own your data."}]},{"type":"paragraph","content":[{"type":"text","text":"Happy writing!"}]}]}`,
			tags:    []string{"welcome", "inkvault"},
		},
		{
			title:   "Why Privacy Matters in Blogging",
			slug:    "why-privacy-matters",
			excerpt: "Most blogging platforms track you. Here's why that's a problem, and what we're doing differently.",
			content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Why Privacy Matters"}]},{"type":"paragraph","content":[{"type":"text","text":"When you write online, you share your thoughts with the world. But most platforms also share your data — with advertisers, data brokers, and analytics companies."}]},{"type":"paragraph","content":[{"type":"text","text":"InkVault is different: we collect the minimum data needed to run the service, encrypt what we do store, and never sell or share your information."}]}]}`,
			tags:    []string{"privacy", "security"},
		},
	}

	for _, p := range posts {
		postID := uuid.New().String()
		_, err = pool.Exec(ctx, `
			INSERT INTO posts
				(id,blog_id,author_id,title,slug,excerpt,content,content_type,
				 status,visibility,is_pinned,tags,reading_time_mins,
				 published_at,created_at,updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,'tiptap','published','public',false,$8,1,$9,$9,$9)`,
			postID, blogID, ownerID,
			p.title, p.slug, p.excerpt, p.content, p.tags, now,
		)
		if err != nil {
			log.Fatalf("seed post %s: %v", p.slug, err)
		}
		fmt.Printf("✅ Post created: /%s → /%s\n", "my-blog", p.slug)
	}

	fmt.Println("\n🌱 Seed complete! Visit http://localhost:3000")
	fmt.Println("   Login: admin / devpassword123")
	fmt.Println("   Blog:  http://localhost:3000/my-blog")
}
