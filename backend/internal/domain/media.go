package domain

import "time"

// Media represents an uploaded file.
type Media struct {
	ID         string    `json:"id"`
	UploaderID string    `json:"uploader_id"`
	BlogID     string    `json:"blog_id,omitempty"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	SizeBytes  int64     `json:"size_bytes"`
	StorageKey string    `json:"storage_key"`
	PublicURL  string    `json:"public_url"`
	CreatedAt  time.Time `json:"created_at"`
}
