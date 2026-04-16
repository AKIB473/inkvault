"use client";

import { useRef, useState } from "react";
import { Image, Upload, X } from "lucide-react";

interface Props {
  blogID: string;
  onUploaded: (url: string) => void;
  label?: string;
}

export function ImageUpload({ blogID, onUploaded, label = "Upload image" }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState("");

  const handleFile = async (file: File) => {
    if (file.size > 10 * 1024 * 1024) {
      setError("File must be under 10MB");
      return;
    }

    setIsUploading(true);
    setError("");

    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("blog_id", blogID);

      const token = sessionStorage.getItem("access_token");
      const res = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/media`,
        {
          method: "POST",
          headers: token ? { Authorization: `Bearer ${token}` } : {},
          credentials: "include",
          body: formData,
        }
      );

      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.message ?? "Upload failed");
      }

      const data = await res.json();
      setPreview(data.url);
      onUploaded(data.url);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div>
      <input
        ref={inputRef}
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        className="hidden"
        onChange={(e) => e.target.files?.[0] && handleFile(e.target.files[0])}
      />

      {preview ? (
        <div className="relative rounded-xl overflow-hidden group">
          <img src={preview} alt="Preview" className="w-full h-48 object-cover" />
          <button
            type="button"
            onClick={() => { setPreview(null); }}
            className="absolute top-2 right-2 p-1.5 rounded-full bg-black/60 text-white opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      ) : (
        <button
          type="button"
          onClick={() => inputRef.current?.click()}
          disabled={isUploading}
          className="w-full h-32 rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-700 flex flex-col items-center justify-center gap-2 hover:border-blue-400 dark:hover:border-blue-600 transition-colors text-gray-400 hover:text-blue-500"
        >
          {isUploading ? (
            <>
              <Upload className="h-5 w-5 animate-bounce" />
              <span className="text-sm">Uploading...</span>
            </>
          ) : (
            <>
              <Image className="h-5 w-5" />
              <span className="text-sm">{label}</span>
              <span className="text-xs text-gray-300">PNG, JPG, GIF, WebP · Max 10MB</span>
            </>
          )}
        </button>
      )}

      {error && <p className="mt-1.5 text-xs text-red-500">{error}</p>}
    </div>
  );
}
