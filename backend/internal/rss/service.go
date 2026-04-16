// Package rss generates RSS 2.0 and Atom feeds for blogs.
package rss

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/you/inkvault/internal/domain"
	"github.com/you/inkvault/internal/repository"
)

type Service struct {
	store   repository.Store
	appURL  string
}

func NewService(store repository.Store, appURL string) *Service {
	return &Service{store: store, appURL: appURL}
}

// GenerateRSS returns an RSS 2.0 feed for the given blog.
func (s *Service) GenerateRSS(ctx context.Context, blogSlug string) (string, error) {
	blog, err := s.store.Blogs().GetBlogBySlug(ctx, blogSlug)
	if err != nil {
		return "", err
	}

	if blog.Visibility != domain.VisibilityPublic {
		return "", fmt.Errorf("feed not available for non-public blog")
	}

	posts, err := s.store.Posts().GetPostsByBlog(ctx, blog.ID, domain.PostStatusPublished, 20, 0)
	if err != nil {
		return "", err
	}

	blogURL := s.blogURL(blog)
	now := time.Now().UTC().Format(time.RFC1123Z)

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n")
	b.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:content="http://purl.org/rss/1.0/modules/content/">`)
	b.WriteString("\n<channel>\n")
	b.WriteString(fmt.Sprintf("  <title>%s</title>\n", xmlEscape(blog.Title)))
	b.WriteString(fmt.Sprintf("  <link>%s</link>\n", blogURL))
	b.WriteString(fmt.Sprintf("  <description>%s</description>\n", xmlEscape(blog.Description)))
	b.WriteString(fmt.Sprintf("  <language>%s</language>\n", blog.Language))
	b.WriteString(fmt.Sprintf("  <lastBuildDate>%s</lastBuildDate>\n", now))
	b.WriteString(fmt.Sprintf(`  <atom:link href="%s/feed.xml" rel="self" type="application/rss+xml"/>`, blogURL))
	b.WriteString("\n")
	b.WriteString("  <generator>InkVault</generator>\n")

	for _, post := range posts {
		if post.Visibility != domain.VisibilityPublic && post.Visibility != domain.VisibilityUnlisted {
			continue
		}
		postURL := fmt.Sprintf("%s/%s", blogURL, post.Slug)
		pubDate := ""
		if post.PublishedAt != nil {
			pubDate = post.PublishedAt.UTC().Format(time.RFC1123Z)
		}

		b.WriteString("  <item>\n")
		b.WriteString(fmt.Sprintf("    <title>%s</title>\n", xmlEscape(post.Title)))
		b.WriteString(fmt.Sprintf("    <link>%s</link>\n", postURL))
		b.WriteString(fmt.Sprintf("    <guid isPermaLink=\"true\">%s</guid>\n", postURL))
		if post.Excerpt != "" {
			b.WriteString(fmt.Sprintf("    <description>%s</description>\n", xmlEscape(post.Excerpt)))
		}
		if pubDate != "" {
			b.WriteString(fmt.Sprintf("    <pubDate>%s</pubDate>\n", pubDate))
		}
		for _, tag := range post.Tags {
			b.WriteString(fmt.Sprintf("    <category>%s</category>\n", xmlEscape(tag)))
		}
		b.WriteString("  </item>\n")
	}

	b.WriteString("</channel>\n</rss>")
	return b.String(), nil
}

// GenerateAtom returns an Atom 1.0 feed.
func (s *Service) GenerateAtom(ctx context.Context, blogSlug string) (string, error) {
	blog, err := s.store.Blogs().GetBlogBySlug(ctx, blogSlug)
	if err != nil {
		return "", err
	}

	posts, err := s.store.Posts().GetPostsByBlog(ctx, blog.ID, domain.PostStatusPublished, 20, 0)
	if err != nil {
		return "", err
	}

	blogURL := s.blogURL(blog)
	updated := time.Now().UTC().Format(time.RFC3339)
	if len(posts) > 0 && posts[0].PublishedAt != nil {
		updated = posts[0].PublishedAt.UTC().Format(time.RFC3339)
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<feed xmlns="http://www.w3.org/2005/Atom" xml:lang="%s">`, blog.Language))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  <title>%s</title>\n", xmlEscape(blog.Title)))
	b.WriteString(fmt.Sprintf("  <subtitle>%s</subtitle>\n", xmlEscape(blog.Description)))
	b.WriteString(fmt.Sprintf("  <link href=\"%s\" rel=\"alternate\"/>\n", blogURL))
	b.WriteString(fmt.Sprintf("  <link href=\"%s/atom.xml\" rel=\"self\"/>\n", blogURL))
	b.WriteString(fmt.Sprintf("  <id>%s</id>\n", blogURL))
	b.WriteString(fmt.Sprintf("  <updated>%s</updated>\n", updated))
	b.WriteString("  <generator>InkVault</generator>\n")

	for _, post := range posts {
		if post.Visibility != domain.VisibilityPublic && post.Visibility != domain.VisibilityUnlisted {
			continue
		}
		postURL := fmt.Sprintf("%s/%s", blogURL, post.Slug)
		entryUpdated := post.UpdatedAt.UTC().Format(time.RFC3339)
		entryPublished := entryUpdated
		if post.PublishedAt != nil {
			entryPublished = post.PublishedAt.UTC().Format(time.RFC3339)
		}

		b.WriteString("  <entry>\n")
		b.WriteString(fmt.Sprintf("    <title>%s</title>\n", xmlEscape(post.Title)))
		b.WriteString(fmt.Sprintf("    <link href=\"%s\"/>\n", postURL))
		b.WriteString(fmt.Sprintf("    <id>%s</id>\n", postURL))
		b.WriteString(fmt.Sprintf("    <published>%s</published>\n", entryPublished))
		b.WriteString(fmt.Sprintf("    <updated>%s</updated>\n", entryUpdated))
		if post.Excerpt != "" {
			b.WriteString(fmt.Sprintf("    <summary>%s</summary>\n", xmlEscape(post.Excerpt)))
		}
		for _, tag := range post.Tags {
			b.WriteString(fmt.Sprintf("    <category term=\"%s\"/>\n", xmlEscape(tag)))
		}
		b.WriteString("  </entry>\n")
	}

	b.WriteString("</feed>")
	return b.String(), nil
}

func (s *Service) blogURL(blog *domain.Blog) string {
	if blog.Domain != "" {
		return "https://" + blog.Domain
	}
	return fmt.Sprintf("%s/%s", s.appURL, blog.Slug)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
