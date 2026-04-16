"use client"

import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { Eye, EyeOff, Mail, User } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

const schema = z.object({
  username: z.string().min(3).max(30).regex(/^[a-zA-Z0-9_]+$/, "Letters, numbers, underscores only"),
  email: z.string().email("Valid email required"),
  password: z.string().min(8).max(128),
  confirm_password: z.string(),
  invite_code: z.string().optional(),
  fullname: z.string().max(0).optional(), // honeypot
}).refine(d => d.password === d.confirm_password, {
  message: "Passwords don't match", path: ["confirm_password"],
})
type Form = z.infer<typeof schema>

export default function RegisterPage() {
  const router = useRouter()
  const [showPw, setShowPw] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")
  const [usernameState, setUsernameState] = useState<"idle"|"checking"|"ok"|"taken">("idle")
  const [timerRef, setTimerRef] = useState<ReturnType<typeof setTimeout>|null>(null)

  const { register, handleSubmit, formState: { errors } } = useForm<Form>({
    resolver: zodResolver(schema),
  })

  const checkUsername = (v: string) => {
    if (timerRef) clearTimeout(timerRef)
    if (v.length < 3) { setUsernameState("idle"); return }
    setUsernameState("checking")
    setTimerRef(setTimeout(async () => {
      try {
        const r = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/auth/username-check?username=${encodeURIComponent(v)}`)
        const j = await r.json()
        setUsernameState(j.available ? "ok" : "taken")
      } catch { setUsernameState("idle") }
    }, 400))
  }

  const onSubmit = async (data: Form) => {
    setLoading(true); setError("")
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/auth/register`, {
        method: "POST", headers: { "Content-Type": "application/json" },
        credentials: "include", body: JSON.stringify(data),
      })
      const json = await res.json()
      if (!res.ok) { setError(json.message || "Registration failed"); return }
      if (json.access_token) sessionStorage.setItem("access_token", json.access_token)
      router.push("/dashboard")
    } catch { setError("Network error. Please try again.") }
    finally { setLoading(false) }
  }

  const usernameIndicator = usernameState === "checking" ? "…"
    : usernameState === "ok" ? <span className="text-green-500 text-xs">✓ Available</span>
    : usernameState === "taken" ? <span className="text-red-500 text-xs">✗ Taken</span>
    : null

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950 flex flex-col items-center justify-center px-4 py-10">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <Link href="/" className="text-3xl">🖋️</Link>
          <h1 className="mt-3 text-2xl font-bold text-gray-900 dark:text-white">Create your account</h1>
          <p className="mt-1 text-sm text-gray-500">Start writing in seconds. No credit card needed.</p>
        </div>

        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-7">
          {/* Honeypot — invisible to humans */}
          <input {...register("fullname")} type="text" tabIndex={-1} aria-hidden="true"
            style={{ position: "absolute", left: -9999, opacity: 0 }} autoComplete="off" />

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            {error && (
              <div className="p-3 rounded-lg bg-red-50 dark:bg-red-950/50 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm">
                {error}
              </div>
            )}

            <div>
              <Input
                {...register("username", { onChange: (e) => checkUsername(e.target.value) })}
                label="Username"
                placeholder="yourname"
                autoComplete="username"
                leftIcon={<User className="h-4 w-4" />}
                error={errors.username?.message || (usernameState === "taken" ? "Username already taken" : undefined)}
              />
              {usernameIndicator && !errors.username && (
                <div className="mt-1">{usernameIndicator}</div>
              )}
            </div>

            <Input
              {...register("email")}
              label="Email"
              type="email"
              placeholder="you@example.com"
              autoComplete="email"
              leftIcon={<Mail className="h-4 w-4" />}
              error={errors.email?.message}
              hint="Encrypted at rest — never stored in plaintext"
            />

            <Input
              {...register("password")}
              label="Password"
              type={showPw ? "text" : "password"}
              placeholder="Min. 8 characters"
              autoComplete="new-password"
              error={errors.password?.message}
              rightIcon={
                <button type="button" onClick={() => setShowPw(v => !v)} className="hover:text-gray-600">
                  {showPw ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              }
            />

            <Input
              {...register("confirm_password")}
              label="Confirm password"
              type="password"
              placeholder="••••••••"
              autoComplete="new-password"
              error={errors.confirm_password?.message}
            />

            <Button type="submit" variant="primary" size="lg" loading={loading} className="w-full mt-2">
              Create account
            </Button>

            <p className="text-xs text-center text-gray-400">
              By signing up you agree to our{" "}
              <Link href="/privacy" className="text-blue-600 hover:underline">Privacy Policy</Link>.
            </p>
          </form>

          <p className="mt-5 text-center text-sm text-gray-500">
            Already have an account?{" "}
            <Link href="/login" className="text-blue-600 hover:underline font-medium">Sign in</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
