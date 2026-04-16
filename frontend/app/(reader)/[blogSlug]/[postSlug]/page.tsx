import { Metadata } from "next";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Clock, Eye, Calendar } from "lucide-react";

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

async function getPost(blogSlug: string, postSlug: string) {
  const res = await fetch(`${BASE}/blogs/${blogSlug}/posts/${postSlug}`, {
    next: { revalidate: 60 },
  });
  if (!res.ok) return null;
  const data = await res.json();
  return data.post;
}

export async function generateMetadata({
  params,
}: {
  params: { blogSlug: string; postSlug: string };
}): Promise<Metadata> {
  const post = await getPost(params.blogSlug, params.postSlug);
  if (!post) return { title: "Post not found" };

  return {
    title: post.seo_title || post.title,
    description: post.seo_description || post.excerpt,
    openGraph: {
      title: post.seo_title || post.title,
      description: post.seo_description || post.excerpt,
      images: post.og_image ? [post.og_image] : post.cover_image ? [post.cover_image] : [],
    },
    // Canonical URL support (Comma pattern)
    alternates: post.canonical_url ? { canonical: post.canonical_url } : undefined,
  };
}

export default async function PostPage({
  params,
}: {
  params: { blogSlug: string; postSlug: string };
}) {
  const post = await getPost(params.blogSlug, params.postSlug);
  if (!post) notFound();

  return (
    <div className="min-h-screen bg-white dark:bg-gray-950">
      <div className="max-w-2xl mx-auto px-4 py-10">
        {/* Back link */}
        <Link
          href={`/${params.blogSlug}`}
          className="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-900 dark:hover:text-white mb-8"
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </Link>

        {/* Cover image */}
        {post.cover_image && (
          <img
            src={post.cover_image}
            alt={post.title}
            className="w-full h-64 object-cover rounded-xl mb-8"
          />
        )}

        {/* Header */}
        <header className="mb-8">
          {post.tags?.length > 0 && (
            <div className="flex gap-2 mb-3">
              {post.tags.map((tag: string) => (
                <span
                  key={tag}
                  className="text-xs px-2.5 py-1 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400 rounded-full"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}

          <h1 className="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white leading-tight">
            {post.title}
          </h1>

          {post.excerpt && (
            <p className="mt-3 text-lg text-gray-600 dark:text-gray-400">{post.excerpt}</p>
          )}

          <div className="flex items-center gap-4 mt-4 text-sm text-gray-400">
            {post.published_at && (
              <span className="flex items-center gap-1.5">
                <Calendar className="h-4 w-4" />
                {new Date(post.published_at).toLocaleDateString("en-US", {
                  year: "numeric", month: "long", day: "numeric",
                })}
              </span>
            )}
            <span className="flex items-center gap-1.5">
              <Clock className="h-4 w-4" />
              {post.reading_time_mins ?? 1} min read
            </span>
            <span className="flex items-center gap-1.5">
              <Eye className="h-4 w-4" />
              {(post.views ?? 0).toLocaleString()} views
            </span>
          </div>
        </header>

        <hr className="border-gray-200 dark:border-gray-800 mb-8" />

        {/* Content */}
        <article
          className="prose prose-gray dark:prose-invert max-w-none
            prose-headings:font-bold
            prose-a:text-blue-600 prose-a:no-underline hover:prose-a:underline
            prose-code:bg-gray-100 dark:prose-code:bg-gray-800 prose-code:rounded prose-code:px-1
            prose-pre:bg-gray-900 prose-pre:text-gray-100"
        >
          {/* Tiptap JSON content rendered as HTML — in production use a Tiptap renderer */}
          {post.content_type === "markdown" ? (
            <div dangerouslySetInnerHTML={{ __html: post.content }} />
          ) : (
            // For Tiptap JSON, use @tiptap/html generateHTML on server side
            <TiptapRenderer content={post.content} />
          )}
        </article>

        {/* Footer */}
        <footer className="mt-12 pt-8 border-t border-gray-200 dark:border-gray-800">
          <Link
            href={`/${params.blogSlug}`}
            className="inline-flex items-center gap-1.5 text-sm text-blue-600 hover:underline"
          >
            <ArrowLeft className="h-4 w-4" />
            More from this blog
          </Link>
        </footer>
      </div>
    </div>
  );
}

// Minimal Tiptap JSON renderer for server-side (no JS needed)
function TiptapRenderer({ content }: { content: string }) {
  try {
    const doc = JSON.parse(content);
    return <TiptapNode node={doc} />;
  } catch {
    return <div>{content}</div>;
  }
}

function TiptapNode({ node }: { node: any }): JSX.Element {
  if (!node) return <></>;

  switch (node.type) {
    case "doc":
      return <>{node.content?.map((n: any, i: number) => <TiptapNode key={i} node={n} />)}</>;
    case "paragraph":
      return <p>{node.content?.map((n: any, i: number) => <TiptapNode key={i} node={n} />)}</p>;
    case "heading":
      const Tag = `h${node.attrs?.level ?? 2}` as keyof JSX.IntrinsicElements;
      return <Tag>{node.content?.map((n: any, i: number) => <TiptapNode key={i} node={n} />)}</Tag>;
    case "text": {
      let el: React.ReactNode = node.text;
      if (node.marks) {
        for (const mark of node.marks) {
          if (mark.type === "bold") el = <strong>{el}</strong>;
          if (mark.type === "italic") el = <em>{el}</em>;
          if (mark.type === "code") el = <code>{el}</code>;
          if (mark.type === "strike") el = <s>{el}</s>;
          if (mark.type === "link") el = <a href={mark.attrs?.href}>{el}</a>;
        }
      }
      return <>{el}</>;
    }
    case "bulletList":
      return <ul>{node.content?.map((n: any, i: number) => <TiptapNode key={i} node={n} />)}</ul>;
    case "orderedList":
      return <ol>{node.content?.map((n: any, i: number) => <TiptapNode key={i} node={n} />)}</ol>;
    case "listItem":
      return <li>{node.content?.map((n: any, i: number) => <TiptapNode key={i} node={n} />)}</li>;
    case "blockquote":
      return <blockquote>{node.content?.map((n: any, i: number) => <TiptapNode key={i} node={n} />)}</blockquote>;
    case "codeBlock":
      return <pre><code className={node.attrs?.language ? `language-${node.attrs.language}` : ""}>{node.content?.map((n: any, i: number) => <TiptapNode key={i} node={n} />)}</code></pre>;
    case "hardBreak":
      return <br />;
    case "horizontalRule":
      return <hr />;
    case "image":
      return <img src={node.attrs?.src} alt={node.attrs?.alt ?? ""} />;
    default:
      return <>{node.content?.map((n: any, i: number) => <TiptapNode key={i} node={n} />) ?? null}</>;
  }
}
