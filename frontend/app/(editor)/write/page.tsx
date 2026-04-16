"use client"

import { useState, useEffect } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import dynamic from "next/dynamic"
import Link from "next/link"
import { useSearchParams, useRouter } from "next/navigation"
import { api } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ArrowLeft, Settings2, Send, Save, ChevronDown, Globe, EyeOff, Eye, Lock } from "lucide-react"

const InkEditor = dynamic(
  () => import("@/components/editor/InkEditor").then(m => m.InkEditor),
  { ssr: false, loading: () => <div className="h-96 rounded-xl bg-gray-100 dark:bg-gray-800 animate-pulse" /> }
)

const schema = z.object({
  blog_id:        z.string().min(1, "Select a blog"),
  title:          z.string().max(500),
  tags:           z.string().optional(),
  visibility:     z.enum(["public","unlisted","members_only","private","password_protected"]),
  seo_title:      z.string().max(255).optional(),
  seo_description:z.string().max(500).optional(),
})
type PostForm = z.infer<typeof schema>

const visOptions = [
  { value: "public",             label: "Public",             icon: Globe,   desc: "Anyone" },
  { value: "unlisted",           label: "Unlisted",           icon: EyeOff,  desc: "Link only" },
  { value: "members_only",       label: "Members",            icon: Eye,     desc: "Logged in" },
  { value: "private",            label: "Private",            icon: Lock,    desc: "Only you" },
] as const

export default function WritePage() {
  const router = useRouter()
  const params = useSearchParams()
  const editID = params.get("id")

  const [content, setContent] = useState("")
  const [blogs, setBlogs] = useState<any[]>([])
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [showSEO, setShowSEO] = useState(false)
  const [wordCount, setWordCount] = useState(0)

  const { register, handleSubmit, watch, setValue, formState: { errors } } = useForm<PostForm>({
    resolver: zodResolver(schema),
    defaultValues: { visibility: "public" },
  })

  const vis = watch("visibility")
  const VisIcon = visOptions.find(v => v.value === vis)?.icon ?? Globe

  useEffect(() => {
    api.blogs.mine().then(r => {
      setBlogs(r.blogs ?? [])
      if (r.blogs?.length) setValue("blog_id", r.blogs[0].id)
    })
    if (editID) {
      // Load existing post for editing
    }
  }, [editID, setValue])

  const handleSave = async (status: "draft" | "published") => {
    setSaving(true)
    try {
      const vals = watch()
      const tags = vals.tags ? vals.tags.split(",").map(t => t.trim()).filter(Boolean) : []
      await api.posts.create({
        blog_id: vals.blog_id,
        title: vals.title || "Untitled",
        content,
        tags,
        visibility: vals.visibility,
        status,
        seo_title: vals.seo_title,
        seo_description: vals.seo_description,
      })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
      if (status === "published") router.push("/dashboard")
    } catch (e) { console.error(e) }
    finally { setSaving(false) }
  }

  return (
    <div className="min-h-screen bg-white dark:bg-gray-950 flex flex-col">
      {/* Toolbar */}
      <header className="sticky top-0 z-50 h-14 border-b border-gray-200 dark:border-gray-800 bg-white/95 dark:bg-gray-950/95 backdrop-blur-sm px-4 flex items-center gap-3">
        <Link href="/dashboard">
          <Button variant="ghost" size="icon-sm"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>

        {/* Blog selector */}
        <select
          {...register("blog_id")}
          className="text-sm border border-gray-200 dark:border-gray-800 rounded-lg px-3 py-1.5 bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 max-w-[160px]"
        >
          {blogs.map(b => <option key={b.id} value={b.id}>{b.title}</option>)}
        </select>

        {/* Visibility */}
        <div className="relative">
          <select
            {...register("visibility")}
            className="appearance-none text-sm border border-gray-200 dark:border-gray-800 rounded-lg pl-8 pr-6 py-1.5 bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            {visOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
          <VisIcon className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-gray-400 pointer-events-none" />
          <ChevronDown className="absolute right-1.5 top-1/2 -translate-y-1/2 h-3 w-3 text-gray-400 pointer-events-none" />
        </div>

        <div className="flex-1" />

        {/* Word count */}
        <span className="hidden sm:block text-xs text-gray-400">{wordCount} words</span>

        {/* SEO toggle */}
        <Button variant="ghost" size="icon-sm" onClick={() => setShowSEO(v => !v)} title="SEO settings">
          <Settings2 className="h-4 w-4" />
        </Button>

        {saved && <Badge variant="success">Saved ✓</Badge>}

        <Button variant="outline" size="sm" loading={saving} onClick={() => handleSave("draft")}>
          <Save className="h-3.5 w-3.5" />
          <span className="hidden sm:block">Save draft</span>
        </Button>

        <Button variant="primary" size="sm" loading={saving} onClick={handleSubmit(() => handleSave("published"))}>
          <Send className="h-3.5 w-3.5" />
          Publish
        </Button>
      </header>

      <div className="flex-1 max-w-3xl w-full mx-auto px-4 py-8 space-y-4">
        {/* Title */}
        <input
          {...register("title")}
          type="text"
          placeholder="Post title…"
          className="w-full text-4xl sm:text-5xl font-bold text-gray-900 dark:text-white bg-transparent border-none outline-none placeholder-gray-200 dark:placeholder-gray-800 leading-tight"
        />

        {/* Tags */}
        <input
          {...register("tags")}
          type="text"
          placeholder="Add tags (comma separated) — e.g. technology, privacy, writing"
          className="w-full text-sm text-gray-400 bg-transparent border-none outline-none placeholder-gray-300 dark:placeholder-gray-700"
        />

        <div className="border-t border-gray-100 dark:border-gray-900" />

        {/* Editor */}
        <InkEditor
          content={content}
          onChange={(c) => {
            setContent(c)
            try {
              const doc = JSON.parse(c)
              const text = extractText(doc)
              setWordCount(text.split(/\s+/).filter(Boolean).length)
            } catch {}
          }}
        />

        {/* SEO panel */}
        {showSEO && (
          <div className="rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 p-5 space-y-4">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-white flex items-center gap-2">
              <Settings2 className="h-4 w-4" /> SEO & Sharing
            </h3>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">SEO title <span className="text-gray-400">(overrides post title in search results)</span></label>
              <input {...register("seo_title")} type="text" placeholder="Leave blank to use post title"
                className="w-full px-3 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">Meta description</label>
              <textarea {...register("seo_description")} rows={2} placeholder="Short description for search engines and social sharing…"
                className="w-full px-3 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function extractText(node: any): string {
  if (!node) return ""
  if (node.type === "text") return node.text ?? ""
  return (node.content ?? []).map(extractText).join(" ")
}
