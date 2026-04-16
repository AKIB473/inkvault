import { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { Rss } from "lucide-react";

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

async function getBlog(slug: string) {
  const res = await fetch(`${BASE}/blogs/${slug}`, { next: { revalidate: 60 } });
  if (!res.ok) return null;
  const data = await res.json();
  return data.blog;
}

async function getBlogPosts(slug: string) {
  const res = await fetch(`${BASE}/blogs/${slug}/posts?limit=20`, { next: { revalidate: 60 } });
  if (!res.ok) return [];
  const data = await res.json();
  return data.posts ?? [];
}

export async function generateMetadata({ params }: { params: { blogSlug: string } }): Promise<Metadata> {
  const blog = await getBlog(params.blogSlug);
  if (!blog) return { title: "Blog not found" };
  return {
    title: blog.title,
    description: blog.description,
    openGraph: { title: blog.title, description: blog.description },
  };
}

export default async function BlogPage({ params }: { params: { blogSlug: string } }) {
  const [blog, posts] = await Promise.all([
    getBlog(params.blogSlug),
    getBlogPosts(params.blogSlug),
  ]);

  if (!blog) notFound();

  return (
    <div className="min-h-screen bg-white dark:bg-gray-950">
      {/* Blog header */}
      <header className="border-b border-gray-200 dark:border-gray-800">
        <div className="max-w-2xl mx-auto px-4 py-12">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">{blog.title}</h1>
          {blog.description && (
            <p className="mt-2 text-gray-600 dark:text-gray-400">{blog.description}</p>
          )}
          <div className="flex items-center gap-4 mt-4">
            <a
              href={`/feed/${params.blogSlug}.xml`}
              className="flex items-center gap-1.5 text-sm text-orange-500 hover:text-orange-600"
              title="RSS Feed"
            >
              <Rss className="h-4 w-4" />
              RSS
            </a>
            <span className="text-sm text-gray-400">{posts.length} posts</span>
          </div>
        </div>
      </header>

      {/* Posts list */}
      <main className="max-w-2xl mx-auto px-4 py-10">
        {posts.length === 0 ? (
          <p className="text-center text-gray-400 py-16">No posts published yet.</p>
        ) : (
          <div className="space-y-10">
            {posts.map((post: any) => (
              <article key={post.id}>
                <Link href={`/${params.blogSlug}/${post.slug}`} className="group block">
                  {post.cover_image && (
                    <img
                      src={post.cover_image}
                      alt={post.title}
                      className="w-full h-48 object-cover rounded-xl mb-4"
                    />
                  )}
                  <h2 className="text-xl font-bold text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
                    {post.title}
                  </h2>
                  {post.excerpt && (
                    <p className="mt-2 text-gray-600 dark:text-gray-400 line-clamp-3">{post.excerpt}</p>
                  )}
                  <div className="flex items-center gap-3 mt-3 text-sm text-gray-400">
                    {post.published_at && (
                      <time dateTime={post.published_at}>
                        {new Date(post.published_at).toLocaleDateString("en-US", {
                          year: "numeric", month: "long", day: "numeric",
                        })}
                      </time>
                    )}
                    <span>·</span>
                    <span>{post.reading_time_mins ?? 1} min read</span>
                    {post.tags?.length > 0 && (
                      <>
                        <span>·</span>
                        <div className="flex gap-1">
                          {post.tags.slice(0, 3).map((tag: string) => (
                            <span key={tag} className="text-xs px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded-full">
                              {tag}
                            </span>
                          ))}
                        </div>
                      </>
                    )}
                  </div>
                </Link>
              </article>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
