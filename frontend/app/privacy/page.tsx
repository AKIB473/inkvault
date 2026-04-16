import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Privacy Policy",
  description: "How InkVault handles your data — the short, honest version.",
};

export default function PrivacyPage() {
  return (
    <div className="min-h-screen bg-white dark:bg-gray-950">
      <div className="max-w-2xl mx-auto px-4 py-12">
        <Link href="/" className="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-900 dark:hover:text-white mb-8">
          <ArrowLeft className="h-4 w-4" /> Home
        </Link>

        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">Privacy Policy</h1>
        <p className="text-gray-500 mb-8">The short, honest version. Last updated: April 2026.</p>

        <div className="prose prose-gray dark:prose-invert max-w-none space-y-8">
          <section>
            <h2>What we collect</h2>
            <ul>
              <li><strong>Username</strong> — to identify your account</li>
              <li><strong>Email address</strong> — encrypted at rest using AES-256-GCM. Never stored in plaintext. Used only for password resets, 2FA codes, and critical notifications.</li>
              <li><strong>Password</strong> — hashed with bcrypt (cost 12). We cannot recover it.</li>
              <li><strong>Content you write</strong> — posts, blogs, drafts. It's yours.</li>
              <li><strong>Technical logs</strong> — IP address, user agent, timestamps on auth events. Kept for 90 days for security purposes.</li>
            </ul>
          </section>

          <section>
            <h2>What we don't collect</h2>
            <ul>
              <li>No advertising IDs or tracking pixels</li>
              <li>No behavioral analytics sent to third parties</li>
              <li>No sale or sharing of your data</li>
              <li>No fingerprinting of your readers</li>
            </ul>
          </section>

          <section>
            <h2>Your rights</h2>
            <ul>
              <li><strong>Export</strong> — Download all your posts and data at any time from Settings → Your Data</li>
              <li><strong>Delete</strong> — Permanently delete your account and all associated data from Settings → Danger Zone. We hard-delete; there is no recovery.</li>
              <li><strong>Correct</strong> — Update your account information any time</li>
            </ul>
          </section>

          <section>
            <h2>Data storage</h2>
            <p>Your data is stored on servers you control (self-hosted) or on infrastructure governed by this policy. Email addresses are encrypted with a key that is never stored in the database.</p>
          </section>

          <section>
            <h2>Cookies</h2>
            <p>We use a single <code>refresh_token</code> cookie: HttpOnly, Secure, SameSite=Strict. It exists solely to keep you logged in. No tracking cookies.</p>
          </section>

          <section>
            <h2>Third-party services</h2>
            <p>If configured, transactional emails may be sent via Resend. Your email address is transmitted to Resend only to deliver that email. Media files are stored in your configured S3-compatible storage.</p>
          </section>

          <section>
            <h2>Contact</h2>
            <p>Questions? Open an issue on GitHub or contact the instance administrator.</p>
          </section>
        </div>
      </div>
    </div>
  );
}
