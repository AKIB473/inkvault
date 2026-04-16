// API client — centralized fetch wrapper with auth token injection.
const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

class APIError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message);
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  // Inject access token from sessionStorage (never localStorage — XSS risk)
  const token = typeof window !== "undefined"
    ? sessionStorage.getItem("access_token")
    : null;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers,
    credentials: "include", // For HttpOnly refresh token cookie
  });

  // Auto-refresh on 401
  if (res.status === 401 && path !== "/auth/refresh") {
    const refreshed = await tryRefresh();
    if (refreshed) {
      return request<T>(path, options); // Retry with new token
    }
    // Redirect to login
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "unknown", message: res.statusText }));
    throw new APIError(res.status, err.error, err.message);
  }

  if (res.status === 204) return {} as T;
  return res.json();
}

async function tryRefresh(): Promise<boolean> {
  try {
    const res = await fetch(`${BASE}/auth/refresh`, {
      method: "POST",
      credentials: "include",
    });
    if (!res.ok) return false;
    const data = await res.json();
    if (data.access_token) {
      sessionStorage.setItem("access_token", data.access_token);
      return true;
    }
    return false;
  } catch {
    return false;
  }
}

// ── API methods ──────────────────────────────────────────────────────────────

export const api = {
  // Auth
  auth: {
    login: (username: string, password: string) =>
      request<{ user: User; access_token: string }>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      }),
    register: (data: RegisterData) =>
      request<{ user: User; access_token: string }>("/auth/register", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    logout: () => request("/auth/logout", { method: "POST" }),
    logoutAll: () => request("/auth/logout-all", { method: "POST" }),
    forgotPassword: (username: string) =>
      request("/auth/forgot-password", {
        method: "POST",
        body: JSON.stringify({ username }),
      }),
    checkUsername: (username: string) =>
      request<UsernameCheckResult>(`/auth/username-check?username=${encodeURIComponent(username)}`),
    providers: () => request<OAuthProviders>("/auth/providers"),
  },

  // Blogs
  blogs: {
    create: (data: CreateBlogData) =>
      request<{ blog: Blog }>("/blogs", { method: "POST", body: JSON.stringify(data) }),
    get: (slug: string) =>
      request<{ blog: Blog }>(`/blogs/${slug}`),
    mine: () =>
      request<{ blogs: Blog[]; total: number }>("/me/blogs"),
  },

  // Posts
  posts: {
    create: (data: CreatePostData) =>
      request<{ post: Post }>("/posts", { method: "POST", body: JSON.stringify(data) }),
    update: (id: string, data: UpdatePostData) =>
      request<{ post: Post }>(`/posts/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/posts/${id}`, { method: "DELETE" }),
    getBySlug: (blogSlug: string, postSlug: string) =>
      request<{ post: Post }>(`/blogs/${blogSlug}/posts/${postSlug}`),
    listByBlog: (blogSlug: string, params?: { limit?: number; offset?: number }) => {
      const qs = new URLSearchParams(params as Record<string, string>).toString();
      return request<{ posts: Post[]; total: number }>(`/blogs/${blogSlug}/posts${qs ? "?" + qs : ""}`);
    },
    mine: (params?: { limit?: number; offset?: number }) => {
      const qs = new URLSearchParams(params as Record<string, string>).toString();
      return request<{ posts: Post[]; total: number }>(`/me/posts${qs ? "?" + qs : ""}`);
    },
    revisions: (id: string) =>
      request<{ revisions: Revision[] }>(`/posts/${id}/revisions`),
    search: (q: string) =>
      request<{ posts: Post[]; query: string }>(`/search?q=${encodeURIComponent(q)}`),
  },

  // Admin
  admin: {
    users: (params?: { limit?: number; offset?: number }) => {
      const qs = new URLSearchParams(params as Record<string, string>).toString();
      return request<{ users: User[] }>(`/admin/users${qs ? "?" + qs : ""}`);
    },
    audit: (params?: { limit?: number; offset?: number }) => {
      const qs = new URLSearchParams(params as Record<string, string>).toString();
      return request<{ logs: AuditLog[] }>(`/admin/audit${qs ? "?" + qs : ""}`);
    },
  },
};

// ── Types ─────────────────────────────────────────────────────────────────────

export interface User {
  id: string;
  username: string;
  display_name: string;
  avatar_url?: string;
  bio?: string;
  created_at: string;
}

export interface Blog {
  id: string;
  owner_id: string;
  slug: string;
  title: string;
  description: string;
  domain?: string;
  visibility: string;
  language: string;
  created_at: string;
  updated_at: string;
  total_posts?: number;
}

export interface Post {
  id: string;
  blog_id: string;
  author_id: string;
  title: string;
  slug: string;
  excerpt: string;
  content: string;
  content_type: string;
  status: "draft" | "scheduled" | "published" | "archived";
  visibility: string;
  is_pinned: boolean;
  tags: string[];
  cover_image?: string;
  seo_title?: string;
  seo_description?: string;
  og_image?: string;
  views: number;
  reading_time_mins: number;
  published_at?: string;
  created_at: string;
  updated_at: string;
  author?: User;
}

export interface Revision {
  id: string;
  post_id: string;
  edited_by: string;
  created_at: string;
}

export interface AuditLog {
  id: string;
  actor_id: string;
  action: string;
  resource_type: string;
  resource_id: string;
  ip_address: string;
  created_at: string;
}

export interface RegisterData {
  username: string;
  email: string;
  password: string;
  confirm_password: string;
  invite_code?: string;
  fullname?: string;
}

export interface CreateBlogData {
  title: string;
  description?: string;
  language?: string;
  visibility?: string;
}

export interface CreatePostData {
  blog_id: string;
  title: string;
  content: string;
  content_type?: string;
  excerpt?: string;
  tags?: string[];
  visibility?: string;
  status?: string;
  cover_image?: string;
  seo_title?: string;
  seo_description?: string;
  og_image?: string;
}

export interface UpdatePostData extends Partial<CreatePostData> {}

export interface UsernameCheckResult {
  available: boolean;
  slug?: string;
  error?: string;
}

export interface OAuthProviders {
  github_enabled: boolean;
  google_enabled: boolean;
  generic_enabled: boolean;
  generic_name?: string;
}
