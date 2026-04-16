"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { api, type Blog, type Post } from "@/lib/api"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { formatRelative } from "@/lib/utils"
import {
  PenLine, BookOpen, Eye, BarChart3, Plus, Settings,
  Globe, EyeOff, Lock, Clock, ArrowUpRight, Loader2,
} from "lucide-react"
import { CreateBlogModal } from "@/components/blog/CreateBlogModal"

const statusBadge: Record<string, { variant: any; label: string }> = {
  published: { variant: "success",   label: "Published" },
  draft:     { variant: "secondary", label: "Draft" },
  scheduled: { variant: "blue",      label: "Scheduled" },
  archived:  { variant: "outline",   label: "Archived" },
}

const visibilityIcon: Record<string, React.ElementType> = {
  public:             Globe,
  unlisted:           EyeOff,
  members_only:       Eye,
  private:            Lock,
  password_protected: Lock,
}

export default function DashboardPage() {
  const [blogs, setBlogs] = useState<Blog[]>([])
  const [posts, setPosts] = useState<Post[]>([])
  const [loading, setLoading] = useState(true)
  const [activeBlog, setActiveBlog] = useState<string | null>(null)
  const [showNewBlog, setShowNewBlog] = useState(false)

  useEffect(() => {
    Promise.all([api.blogs.mine(), api.posts.mine({ limit: 50 })])
      .then(([b, p]) => {
        setBlogs(b.blogs ?? [])
        setPosts(p.posts ?? [])
        if (b.blogs?.length) setActiveBlog(b.blogs[0].id)
      })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [])

  const filtered = activeBlog ? posts.filter(p => p.blog_id === activeBlog) : posts
  const stats = {
    total: posts.length,
    published: posts.filter(p => p.status === "published").length,
    drafts: posts.filter(p => p.status === "draft").length,
    views: posts.reduce((s, p) => s + (p.views || 0), 0),
  }

  if (loading) return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950">
      <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
    </div>
  )

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950">
      {showNewBlog && (
        <CreateBlogModal
          onCreated={(b) => { setBlogs(prev => [b, ...prev]); setActiveBlog(b.id); setShowNewBlog(false) }}
          onClose={() => setShowNewBlog(false)}
        />
      )}

      {/* Topnav */}
      <header className="sticky top-0 z-40 h-14 border-b border-gray-200 dark:border-gray-800 bg-white/90 dark:bg-gray-900/90 backdrop-blur-sm px-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link href="/" className="text-xl font-bold flex items-center gap-1.5">
            <span>🖋️</span><span className="hidden sm:block">InkVault</span>
          </Link>
          <span className="hidden sm:block text-gray-300 dark:text-gray-700">/</span>
          <span className="hidden sm:block text-sm text-gray-500">Dashboard</span>
        </div>
        <div className="flex items-center gap-2">
          <Link href="/write">
            <Button variant="primary" size="sm">
              <PenLine className="h-3.5 w-3.5" />
              <span className="hidden sm:block">New post</span>
            </Button>
          </Link>
          <Link href="/settings">
            <Button variant="ghost" size="icon"><Settings className="h-4 w-4" /></Button>
          </Link>
        </div>
      </header>

      <div className="max-w-6xl mx-auto px-4 py-8 space-y-7">
        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[
            { icon: BookOpen, label: "Total posts",   value: stats.total,                color: "text-blue-600 bg-blue-50 dark:bg-blue-950/40" },
            { icon: Globe,    label: "Published",     value: stats.published,            color: "text-green-600 bg-green-50 dark:bg-green-950/40" },
            { icon: PenLine,  label: "Drafts",        value: stats.drafts,               color: "text-orange-600 bg-orange-50 dark:bg-orange-950/40" },
            { icon: BarChart3,label: "Total views",   value: stats.views.toLocaleString(), color: "text-purple-600 bg-purple-50 dark:bg-purple-950/40" },
          ].map(s => (
            <Card key={s.label} className="p-5">
              <div className={`inline-flex p-2 rounded-lg mb-3 ${s.color}`}>
                <s.icon className="h-4 w-4" />
              </div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">{s.value}</div>
              <div className="text-xs text-gray-500 mt-0.5">{s.label}</div>
            </Card>
          ))}
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Sidebar — blogs */}
          <div className="space-y-2">
            <div className="flex items-center justify-between px-1 mb-3">
              <span className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Blogs</span>
              <Button variant="ghost" size="icon-sm" onClick={() => setShowNewBlog(true)}>
                <Plus className="h-3.5 w-3.5" />
              </Button>
            </div>

            {blogs.length === 0 ? (
              <button
                onClick={() => setShowNewBlog(true)}
                className="w-full flex flex-col items-center justify-center gap-2 h-24 rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-700 text-gray-400 hover:border-blue-400 hover:text-blue-500 transition-colors text-sm"
              >
                <Plus className="h-5 w-5" />
                Create first blog
              </button>
            ) : (
              blogs.map(blog => (
                <button
                  key={blog.id}
                  onClick={() => setActiveBlog(blog.id)}
                  className={`w-full text-left px-3 py-2.5 rounded-lg text-sm transition-colors border ${
                    activeBlog === blog.id
                      ? "bg-blue-50 dark:bg-blue-950/30 border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-400"
                      : "border-transparent hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300"
                  }`}
                >
                  <div className="font-medium truncate">{blog.title}</div>
                  <div className="text-xs text-gray-400 mt-0.5 truncate">/{blog.slug}</div>
                </button>
              ))
            )}

            {activeBlog && (
              <div className="pt-2 space-y-1">
                <a
                  href={`/${blogs.find(b => b.id === activeBlog)?.slug}`}
                  target="_blank"
                  rel="noopener"
                  className="flex items-center gap-1.5 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 px-1"
                >
                  <ArrowUpRight className="h-3 w-3" /> View blog
                </a>
              </div>
            )}
          </div>

          {/* Main — posts */}
          <div className="lg:col-span-3 space-y-3">
            <div className="flex items-center justify-between px-1">
              <span className="text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Posts {activeBlog && blogs.find(b=>b.id===activeBlog) ? `— ${blogs.find(b=>b.id===activeBlog)!.title}` : ""}
              </span>
              <Link href="/write">
                <Button variant="ghost" size="sm" className="text-xs">
                  <Plus className="h-3 w-3" /> Write
                </Button>
              </Link>
            </div>

            {filtered.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-20 rounded-2xl border-2 border-dashed border-gray-200 dark:border-gray-800">
                <PenLine className="h-8 w-8 text-gray-300 dark:text-gray-700 mb-3" />
                <p className="text-sm text-gray-500 mb-4">No posts yet</p>
                <Link href="/write"><Button variant="primary" size="sm">Write your first post</Button></Link>
              </div>
            ) : (
              filtered.map(post => {
                const sb = statusBadge[post.status] ?? statusBadge.draft
                const VisIcon = visibilityIcon[post.visibility] ?? Globe
                return (
                  <Card key={post.id} className="p-4 hover:border-gray-300 dark:hover:border-gray-700 transition-colors">
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex-1 min-w-0">
                        <div className="flex flex-wrap items-center gap-2 mb-1.5">
                          <Badge variant={sb.variant}>{sb.label}</Badge>
                          <VisIcon className="h-3.5 w-3.5 text-gray-400" />
                          {post.tags?.slice(0,3).map(t => (
                            <span key={t} className="text-xs text-gray-400">#{t}</span>
                          ))}
                        </div>
                        <h3 className="font-semibold text-gray-900 dark:text-white truncate">
                          {post.title || <span className="text-gray-400 italic font-normal">Untitled</span>}
                        </h3>
                        {post.excerpt && (
                          <p className="text-sm text-gray-500 mt-1 line-clamp-1">{post.excerpt}</p>
                        )}
                        <div className="flex items-center gap-3 mt-2 text-xs text-gray-400">
                          <span className="flex items-center gap-1"><Eye className="h-3 w-3"/>{post.views}</span>
                          <span className="flex items-center gap-1"><Clock className="h-3 w-3"/>{post.reading_time_mins}m</span>
                          <span className="ml-auto">{formatRelative(post.updated_at)}</span>
                        </div>
                      </div>
                      <Link href={`/write?id=${post.id}`}>
                        <Button variant="outline" size="sm" className="shrink-0 text-xs">Edit</Button>
                      </Link>
                    </div>
                  </Card>
                )
              })
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
