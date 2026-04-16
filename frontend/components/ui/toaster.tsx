"use client"

// Minimal toast notification system
import { useState, useEffect, createContext, useContext, useCallback } from "react"
import { cn } from "@/lib/utils"
import { X } from "lucide-react"

type Toast = { id: string; message: string; type: "success" | "error" | "info" }
type ToastCtx = { toast: (msg: string, type?: Toast["type"]) => void }

const ctx = createContext<ToastCtx>({ toast: () => {} })
export const useToast = () => useContext(ctx)

export function Toaster() {
  const [toasts, setToasts] = useState<Toast[]>([])

  const toast = useCallback((message: string, type: Toast["type"] = "info") => {
    const id = Math.random().toString(36).slice(2)
    setToasts(p => [...p, { id, message, type }])
    setTimeout(() => setToasts(p => p.filter(t => t.id !== id)), 4000)
  }, [])

  const remove = (id: string) => setToasts(p => p.filter(t => t.id !== id))

  return (
    <ctx.Provider value={{ toast }}>
      <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 max-w-sm w-full">
        {toasts.map(t => (
          <div key={t.id} className={cn(
            "flex items-center gap-3 px-4 py-3 rounded-xl shadow-lg border text-sm font-medium animate-in slide-in-from-right-full",
            t.type === "success" && "bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800 text-green-800 dark:text-green-200",
            t.type === "error"   && "bg-red-50 dark:bg-red-950 border-red-200 dark:border-red-800 text-red-800 dark:text-red-200",
            t.type === "info"    && "bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-800 dark:text-gray-200",
          )}>
            <span className="flex-1">{t.message}</span>
            <button onClick={() => remove(t.id)} className="opacity-50 hover:opacity-100">
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        ))}
      </div>
    </ctx.Provider>
  )
}
