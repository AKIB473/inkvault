"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { api } from "@/lib/api";
import { Shield, Lock, Download, Trash2, ArrowLeft, Eye, EyeOff } from "lucide-react";

const passwordSchema = z.object({
  current_password: z.string().min(1),
  new_password: z.string().min(8).max(128),
  confirm_password: z.string(),
}).refine(d => d.new_password === d.confirm_password, {
  message: "Passwords don't match", path: ["confirm_password"],
});

type PasswordForm = z.infer<typeof passwordSchema>;

export default function SettingsPage() {
  const [showCurrent, setShowCurrent] = useState(false);
  const [showNew, setShowNew] = useState(false);
  const [pwSuccess, setPwSuccess] = useState("");
  const [pwError, setPwError] = useState("");
  const [isChangingPw, setIsChangingPw] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const { register, handleSubmit, reset, formState: { errors } } = useForm<PasswordForm>({
    resolver: zodResolver(passwordSchema),
  });

  const onChangePassword = async (data: PasswordForm) => {
    setIsChangingPw(true);
    setPwError("");
    setPwSuccess("");
    try {
      // TODO: wire to PATCH /me/password endpoint
      await new Promise(r => setTimeout(r, 500));
      setPwSuccess("Password changed successfully. All other sessions have been revoked.");
      reset();
    } catch (e: any) {
      setPwError(e.message ?? "Failed to change password");
    } finally {
      setIsChangingPw(false);
    }
  };

  const exportData = async () => {
    const token = sessionStorage.getItem("access_token");
    const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/me/export/posts.json`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "inkvault-export.json";
    a.click();
    URL.revokeObjectURL(url);
  };

  const logoutAll = async () => {
    await api.auth.logoutAll();
    window.location.href = "/login";
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950">
      <div className="max-w-2xl mx-auto px-4 py-10">
        <Link href="/dashboard" className="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-900 dark:hover:text-white mb-8">
          <ArrowLeft className="h-4 w-4" />
          Dashboard
        </Link>

        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-8">Settings</h1>

        {/* Password section */}
        <section className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-6 mb-4">
          <div className="flex items-center gap-2 mb-4">
            <Lock className="h-4 w-4 text-gray-500" />
            <h2 className="font-semibold text-gray-900 dark:text-white">Change Password</h2>
          </div>

          <form onSubmit={handleSubmit(onChangePassword)} className="space-y-4">
            {pwSuccess && <div className="p-3 rounded-lg bg-green-50 dark:bg-green-950 text-green-700 dark:text-green-400 text-sm">{pwSuccess}</div>}
            {pwError && <div className="p-3 rounded-lg bg-red-50 dark:bg-red-950 text-red-700 dark:text-red-400 text-sm">{pwError}</div>}

            {[
              { name: "current_password" as const, label: "Current password", show: showCurrent, toggle: () => setShowCurrent(v => !v) },
              { name: "new_password" as const, label: "New password", show: showNew, toggle: () => setShowNew(v => !v) },
              { name: "confirm_password" as const, label: "Confirm new password", show: showNew, toggle: () => {} },
            ].map(({ name, label, show, toggle }) => (
              <div key={name}>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">{label}</label>
                <div className="relative">
                  <input
                    {...register(name)}
                    type={show ? "text" : "password"}
                    className="w-full pl-3 pr-10 py-2.5 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  {name !== "confirm_password" && (
                    <button type="button" onClick={toggle} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400">
                      {show ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  )}
                </div>
                {errors[name] && <p className="mt-1 text-xs text-red-500">{errors[name]?.message}</p>}
              </div>
            ))}

            <button
              type="submit"
              disabled={isChangingPw}
              className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white text-sm font-medium"
            >
              {isChangingPw ? "Updating..." : "Update password"}
            </button>
          </form>
        </section>

        {/* Security section */}
        <section className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-6 mb-4">
          <div className="flex items-center gap-2 mb-4">
            <Shield className="h-4 w-4 text-gray-500" />
            <h2 className="font-semibold text-gray-900 dark:text-white">Security</h2>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between py-2 border-b border-gray-100 dark:border-gray-800">
              <div>
                <div className="text-sm font-medium text-gray-900 dark:text-white">Two-factor authentication</div>
                <div className="text-xs text-gray-500">Add an extra layer of security to your account</div>
              </div>
              <button className="px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-700 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800">
                Enable
              </button>
            </div>
            <div className="flex items-center justify-between py-2">
              <div>
                <div className="text-sm font-medium text-gray-900 dark:text-white">Sign out all devices</div>
                <div className="text-xs text-gray-500">Revoke all active sessions everywhere</div>
              </div>
              <button
                onClick={logoutAll}
                className="px-3 py-1.5 rounded-lg border border-orange-300 dark:border-orange-800 text-xs text-orange-700 dark:text-orange-400 hover:bg-orange-50 dark:hover:bg-orange-950"
              >
                Sign out all
              </button>
            </div>
          </div>
        </section>

        {/* Privacy / Data section */}
        <section className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-6 mb-4">
          <div className="flex items-center gap-2 mb-4">
            <Download className="h-4 w-4 text-gray-500" />
            <h2 className="font-semibold text-gray-900 dark:text-white">Your Data</h2>
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
            You own your content. Export it any time, in any format.
          </p>
          <div className="flex gap-2">
            <button
              onClick={exportData}
              className="px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-700 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              Export posts (JSON)
            </button>
          </div>
        </section>

        {/* Danger zone */}
        <section className="bg-white dark:bg-gray-900 rounded-2xl border border-red-200 dark:border-red-900 p-6">
          <div className="flex items-center gap-2 mb-4">
            <Trash2 className="h-4 w-4 text-red-500" />
            <h2 className="font-semibold text-red-600 dark:text-red-400">Danger Zone</h2>
          </div>

          {!showDeleteConfirm ? (
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm font-medium text-gray-900 dark:text-white">Delete account</div>
                <div className="text-xs text-gray-500">Permanently delete your account and all data. This cannot be undone.</div>
              </div>
              <button
                onClick={() => setShowDeleteConfirm(true)}
                className="px-3 py-1.5 rounded-lg border border-red-300 dark:border-red-800 text-xs text-red-700 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950"
              >
                Delete account
              </button>
            </div>
          ) : (
            <div className="space-y-3">
              <p className="text-sm text-red-600 dark:text-red-400 font-medium">
                Are you absolutely sure? All your blogs, posts, and data will be permanently deleted.
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setShowDeleteConfirm(false)}
                  className="flex-1 py-2 rounded-lg border border-gray-300 dark:border-gray-700 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={() => {
                    // TODO: call DELETE /me with confirmation token
                    alert("Account deletion requires email confirmation. Check your inbox.");
                  }}
                  className="flex-1 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white text-sm font-medium"
                >
                  Yes, delete everything
                </button>
              </div>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
