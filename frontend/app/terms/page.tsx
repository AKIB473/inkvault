import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Terms of Service",
  description: "InkVault Terms of Service — what you agree to when you use the platform.",
};

export default function TermsPage() {
  return (
    <main className="mx-auto max-w-3xl px-4 py-16">
      <div className="prose dark:prose-invert max-w-none">
        <h1>Terms of Service</h1>
        <p className="lead text-gray-500 dark:text-gray-400">
          Last updated: April 2026
        </p>

        <h2>1. Acceptance of Terms</h2>
        <p>
          By accessing or using InkVault, you agree to be bound by these Terms of Service.
          If you do not agree, do not use the platform.
        </p>

        <h2>2. Your Content</h2>
        <p>
          You retain full ownership of the content you publish on InkVault. By posting, you
          grant InkVault a limited, non-exclusive licence to display and deliver your content
          to your readers.
        </p>

        <h2>3. Prohibited Conduct</h2>
        <p>You agree not to:</p>
        <ul>
          <li>Post content that is illegal, harmful, or harasses others</li>
          <li>Attempt to gain unauthorised access to the platform</li>
          <li>Use the platform to distribute spam or malware</li>
          <li>Violate any applicable laws or regulations</li>
        </ul>

        <h2>4. Privacy</h2>
        <p>
          Your email address is encrypted at rest using AES-256-GCM. We do not sell or share
          your personal data with third parties for marketing purposes.
          See our Privacy Policy for full details.
        </p>

        <h2>5. Service Availability</h2>
        <p>
          InkVault is provided &ldquo;as is.&rdquo; We make reasonable efforts to maintain
          uptime but do not guarantee uninterrupted access.
        </p>

        <h2>6. Termination</h2>
        <p>
          We reserve the right to suspend or terminate accounts that violate these terms,
          with or without notice.
        </p>

        <h2>7. Changes to Terms</h2>
        <p>
          We may update these terms from time to time. Continued use of InkVault after changes
          constitutes acceptance of the new terms.
        </p>

        <h2>8. Contact</h2>
        <p>
          Questions? Reach out via the{" "}
          <Link href="/contact" className="text-blue-600 hover:underline">
            contact page
          </Link>
          .
        </p>
      </div>
    </main>
  );
}
