"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import Image from "@tiptap/extension-image";
import Link from "@tiptap/extension-link";
import CharacterCount from "@tiptap/extension-character-count";
import { EditorToolbar } from "./EditorToolbar";

interface InkEditorProps {
  content?: string;
  onChange?: (json: string) => void;
  placeholder?: string;
  editable?: boolean;
}

export function InkEditor({
  content,
  onChange,
  placeholder = "Start writing your story...",
  editable = true,
}: InkEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: { levels: [1, 2, 3] },
        codeBlock: { languageClassPrefix: "language-" },
      }),
      Placeholder.configure({ placeholder }),
      Image.configure({ inline: false, allowBase64: false }),
      Link.configure({
        openOnClick: false,
        autolink: true,
        protocols: ["https", "http", "mailto"],
      }),
      CharacterCount,
    ],
    content: content ? JSON.parse(content) : "",
    editable,
    onUpdate: ({ editor }) => {
      const json = JSON.stringify(editor.getJSON());
      onChange?.(json);
    },
    editorProps: {
      attributes: {
        class: "prose prose-gray dark:prose-invert max-w-none min-h-[400px] focus:outline-none px-1",
      },
    },
  });

  if (!editor) return null;

  return (
    <div className="border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden">
      {editable && <EditorToolbar editor={editor} />}
      <div className="p-6">
        <EditorContent editor={editor} />
      </div>
      {editable && (
        <div className="px-6 py-2 border-t border-gray-100 dark:border-gray-800 text-xs text-gray-400 flex justify-between">
          <span>{editor.storage.characterCount.words()} words</span>
          <span>~{Math.ceil(editor.storage.characterCount.words() / 200)} min read</span>
        </div>
      )}
    </div>
  );
}
