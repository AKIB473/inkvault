package media

import (
	"testing"
)

func TestSniffMIME(t *testing.T) {
	cases := []struct {
		name    string
		header  []byte
		want    string
	}{
		{"JPEG", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "image/jpeg"},
		{"PNG",  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "image/png"},
		{"GIF",  []byte{0x47, 0x49, 0x46, 0x38}, "image/gif"},
		{"PDF",  []byte{0x25, 0x50, 0x44, 0x46, 0x2D}, "application/pdf"},
		{"unknown", []byte{0x00, 0x01, 0x02, 0x03}, "application/octet-stream"},
	}

	for _, tc := range cases {
		got := sniffMIME(tc.header)
		if got != tc.want {
			t.Errorf("[%s] sniffMIME = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestMimeToExt(t *testing.T) {
	cases := []struct {
		mime string
		want string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"application/pdf", ".pdf"},
		{"unknown/type", ".bin"},
	}

	for _, tc := range cases {
		got := mimeToExt(tc.mime)
		if got != tc.want {
			t.Errorf("mimeToExt(%q) = %q, want %q", tc.mime, got, tc.want)
		}
	}
}

func TestAllowedMIME(t *testing.T) {
	allowed := []string{"image/jpeg", "image/png", "image/gif", "image/webp", "application/pdf"}
	blocked := []string{"application/javascript", "text/html", "application/x-executable", "video/mp4"}

	for _, m := range allowed {
		if !AllowedMIME[m] {
			t.Errorf("%q should be allowed", m)
		}
	}
	for _, m := range blocked {
		if AllowedMIME[m] {
			t.Errorf("%q should NOT be allowed", m)
		}
	}
}

func TestUploadSizeLimit(t *testing.T) {
	// Simulate oversized data
	data := make([]byte, 10<<20+1) // 10MB + 1 byte
	if int64(len(data)) <= maxUploadBytes {
		t.Error("test data should exceed limit")
	}
}
