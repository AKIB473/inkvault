import Link from "next/link"
import { ArrowRight, Shield, Lock, FileText, Rss, Globe, Search, Users, Zap, GitBranch } from "lucide-react"
import { Button } from "@/components/ui/button"

export default function HomePage() {
  return (
    <div className="min-h-screen bg-white dark:bg-gray-950 text-gray-900 dark:text-white">
      {/* Nav */}
      <nav className="fixed top-0 inset-x-0 z-50 border-b border-gray-200/80 dark:border-gray-800/80 bg-white/80 dark:bg-gray-950/80 backdrop-blur-md">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
          <Link href="/" className="flex items-center gap-2 font-bold text-lg">
            <span className="text-2xl">🖋️</span>
            <span>InkVault</span>
          </Link>
          <div className="hidden md:flex items-center gap-6 text-sm text-gray-600 dark:text-gray-400">
            <Link href="#features" className="hover:text-gray-900 dark:hover:text-white transition-colors">Features</Link>
            <Link href="#security" className="hover:text-gray-900 dark:hover:text-white transition-colors">Security</Link>
            <Link href="/privacy" className="hover:text-gray-900 dark:hover:text-white transition-colors">Privacy</Link>
            <a href="https://github.com/your-username/inkvault" target="_blank" rel="noopener" className="hover:text-gray-900 dark:hover:text-white transition-colors">GitHub ↗</a>
          </div>
          <div className="flex items-center gap-2">
            <Link href="/login">
              <Button variant="ghost" size="sm">Sign in</Button>
            </Link>
            <Link href="/register">
              <Button variant="primary" size="sm">Get started</Button>
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero */}
      <section className="pt-32 pb-20 px-4 text-center">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-400 text-xs font-medium mb-6">
          <span className="w-1.5 h-1.5 rounded-full bg-blue-500 animate-pulse"/>
          Open source · Self-hostable · Privacy-first
        </div>

        <h1 className="text-5xl sm:text-6xl lg:text-7xl font-bold tracking-tight leading-[1.08] max-w-3xl mx-auto">
          Write freely.<br/>
          <span className="bg-gradient-to-r from-blue-600 to-violet-600 bg-clip-text text-transparent">
            Own everything.
          </span>
        </h1>

        <p className="mt-6 text-lg sm:text-xl text-gray-500 dark:text-gray-400 max-w-xl mx-auto leading-relaxed">
          InkVault is a secure, encrypted blog platform built for writers who care about privacy.
          No trackers, no ads, no lock-in.
        </p>

        <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-3">
          <Link href="/register">
            <Button variant="primary" size="xl" className="w-full sm:w-auto">
              Start writing for free
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
          <a href="https://github.com/your-username/inkvault" target="_blank" rel="noopener">
            <Button variant="outline" size="xl" className="w-full sm:w-auto gap-2">
              <svg viewBox="0 0 24 24" className="h-4 w-4 fill-current"><path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/></svg>
              View on GitHub
            </Button>
          </a>
        </div>

        {/* Social proof */}
        <p className="mt-8 text-sm text-gray-400">
          Research-backed from{" "}
          <span className="font-medium text-gray-600 dark:text-gray-400">Ghost</span>{" "}(52k ⭐),{" "}
          <span className="font-medium text-gray-600 dark:text-gray-400">WriteFreely</span>{" "}(5.1k ⭐), and{" "}
          <span className="font-medium text-gray-600 dark:text-gray-400">Comma</span>{" "}(218 ⭐)
        </p>
      </section>

      {/* Feature grid */}
      <section id="features" className="py-20 px-4 bg-gray-50 dark:bg-gray-900/50">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-14">
            <h2 className="text-3xl sm:text-4xl font-bold">Everything you need to publish</h2>
            <p className="mt-3 text-gray-500 dark:text-gray-400 max-w-lg mx-auto">
              Professional writing tools with privacy built in from day one — not bolted on.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {[
              {
                icon: FileText, color: "text-blue-600 bg-blue-50 dark:bg-blue-950/50",
                title: "Rich text editor",
                desc: "Notion-style Tiptap editor with Markdown support, code blocks, embeds, and image uploads.",
              },
              {
                icon: Shield, color: "text-violet-600 bg-violet-50 dark:bg-violet-950/50",
                title: "Encrypted by default",
                desc: "Emails encrypted with AES-256-GCM at rest. Passwords hashed with bcrypt cost-12. Zero plaintext PII.",
              },
              {
                icon: Globe, color: "text-green-600 bg-green-50 dark:bg-green-950/50",
                title: "Multiple blogs",
                desc: "Create unlimited blogs per account. Custom domains, pen names, and independent branding.",
              },
              {
                icon: Lock, color: "text-orange-600 bg-orange-50 dark:bg-orange-950/50",
                title: "Granular visibility",
                desc: "Public, unlisted, members-only, private, or password-protected posts. Per-post and per-blog control.",
              },
              {
                icon: Rss, color: "text-rose-600 bg-rose-50 dark:bg-rose-950/50",
                title: "RSS & Atom feeds",
                desc: "Auto-generated feeds for every blog. Let readers subscribe without social media.",
              },
              {
                icon: Search, color: "text-cyan-600 bg-cyan-50 dark:bg-cyan-950/50",
                title: "Full-text search",
                desc: "PostgreSQL FTS with ranking. Searches title, excerpt, and content in milliseconds.",
              },
              {
                icon: GitBranch, color: "text-teal-600 bg-teal-50 dark:bg-teal-950/50",
                title: "Revision history",
                desc: "Every save creates a revision. Roll back to any previous version at any time.",
              },
              {
                icon: Users, color: "text-indigo-600 bg-indigo-50 dark:bg-indigo-950/50",
                title: "Role-based access",
                desc: "Owner → Admin → Editor → Writer → Reader. Invite-only registration option.",
              },
              {
                icon: Zap, color: "text-yellow-600 bg-yellow-50 dark:bg-yellow-950/50",
                title: "High performance",
                desc: "Go backend, Redis caching, Postgres FTS. Sub-10ms response times in production.",
              },
            ].map((f) => (
              <div key={f.title} className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-6 hover:border-gray-300 dark:hover:border-gray-700 transition-colors">
                <div className={`inline-flex p-2.5 rounded-xl mb-4 ${f.color}`}>
                  <f.icon className="h-5 w-5" />
                </div>
                <h3 className="font-semibold text-gray-900 dark:text-white mb-1.5">{f.title}</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 leading-relaxed">{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Security deep-dive */}
      <section id="security" className="py-20 px-4">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-14">
            <h2 className="text-3xl sm:text-4xl font-bold">Security you can audit</h2>
            <p className="mt-3 text-gray-500 dark:text-gray-400">
              Open source. Every security decision documented and reviewable.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[
              { label: "Password hashing", value: "bcrypt cost-12 (OWASP)" },
              { label: "Email encryption", value: "AES-256-GCM at rest" },
              { label: "Session tokens", value: "JWT 15min + rotating refresh (7d)" },
              { label: "Token storage", value: "HttpOnly Secure SameSite=Strict cookie" },
              { label: "Rate limiting", value: "5 auth attempts / hour / IP (Redis)" },
              { label: "Reset tokens", value: "HMAC-signed, embeds password hash" },
              { label: "Spam prevention", value: "Honeypot + invite-only mode" },
              { label: "Audit trail", value: "Every write logged with actor + IP" },
              { label: "GDPR export", value: "One-click JSON export" },
              { label: "Right to deletion", value: "Hard delete with FK cascade" },
              { label: "HTTP headers", value: "HSTS, CSP, X-Frame-Options, no-sniff" },
              { label: "Docker", value: "Runs as uid 65534 (nobody), scratch image" },
            ].map((item) => (
              <div key={item.label} className="flex items-center justify-between p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900">
                <span className="text-sm text-gray-600 dark:text-gray-400">{item.label}</span>
                <span className="text-sm font-mono font-medium text-gray-900 dark:text-white text-right max-w-[55%]">{item.value}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Stack */}
      <section className="py-16 px-4 bg-gray-50 dark:bg-gray-900/50 border-t border-gray-200 dark:border-gray-800">
        <div className="max-w-4xl mx-auto text-center">
          <p className="text-sm font-medium text-gray-400 uppercase tracking-wider mb-6">Built with</p>
          <div className="flex flex-wrap justify-center gap-3">
            {["Go 1.22", "Fiber v2", "Next.js 14", "TypeScript", "PostgreSQL 16", "Redis 7", "Tiptap v2", "Tailwind CSS", "Docker", "GitHub Actions"].map((t) => (
              <span key={t} className="px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 text-sm text-gray-600 dark:text-gray-400 font-medium">
                {t}
              </span>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-24 px-4">
        <div className="max-w-2xl mx-auto text-center">
          <h2 className="text-4xl font-bold mb-4">Ready to own your writing?</h2>
          <p className="text-gray-500 dark:text-gray-400 mb-8 text-lg">
            Free to self-host. Your data stays yours — always.
          </p>
          <Link href="/register">
            <Button variant="primary" size="xl">
              Create your account
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
          <p className="mt-4 text-sm text-gray-400">
            No credit card. No tracking. <Link href="/privacy" className="underline hover:text-gray-600">Privacy policy →</Link>
          </p>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-gray-200 dark:border-gray-800 py-10 px-4">
        <div className="max-w-6xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-gray-400">
          <div className="flex items-center gap-2">
            <span className="text-lg">🖋️</span>
            <span className="font-semibold text-gray-600 dark:text-gray-400">InkVault</span>
            <span>·</span>
            <span>Open source blog platform</span>
          </div>
          <div className="flex items-center gap-5">
            <Link href="/privacy" className="hover:text-gray-600 dark:hover:text-gray-200 transition-colors">Privacy</Link>
            <Link href="/terms" className="hover:text-gray-600 dark:hover:text-gray-200 transition-colors">Terms</Link>
            <a href="https://github.com/your-username/inkvault" target="_blank" rel="noopener" className="hover:text-gray-600 dark:hover:text-gray-200 transition-colors">GitHub</a>
          </div>
        </div>
      </footer>
    </div>
  )
}
