# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest  | ✅        |

## Reporting a Vulnerability

**Please do NOT open a public GitHub issue for security vulnerabilities.**

Report security issues privately via GitHub's [Security Advisories](../../security/advisories/new) feature.

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (optional)

We'll respond within 48 hours and aim to patch within 7 days of confirmation.

## Security Architecture

- **Passwords**: bcrypt cost 12 (OWASP compliant)
- **Emails**: AES-256-GCM encrypted at rest — never stored in plaintext
- **Tokens**: HMAC-SHA256 signed, single-use, time-limited
- **Sessions**: Refresh tokens stored as SHA-256 hashes only
- **Auth**: JWT (15min) + HttpOnly cookie refresh tokens (7 days, rotating)
- **Rate limiting**: Redis-backed, 5 auth attempts/hr/IP
- **Headers**: HSTS, CSP, X-Frame-Options, Referrer-Policy on all responses
- **Process**: Runs as non-root in Docker (uid 65534)
- **Dependencies**: Scanned in CI via Trivy

## Known Non-Issues

- XSS via content: Authors are trusted users — rich text editor output is stored as Tiptap JSON, not raw HTML. Server-side rendering uses a structured renderer, not `dangerouslySetInnerHTML` with raw content.
