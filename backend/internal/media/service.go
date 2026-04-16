// Package media handles file uploads to S3-compatible storage (MinIO).
package media

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/you/inkvault/internal/config"
)

// AllowedMIME is the set of allowed upload MIME types.
var AllowedMIME = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"image/svg+xml":   true,
	"application/pdf": true,
}

const maxUploadBytes = 10 << 20 // 10 MB

type Service struct {
	s3     *s3.Client
	bucket string
	pubURL string
}

func NewService(cfg *config.Config) (*Service, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{URL: cfg.S3Endpoint, HostnameImmutable: true}, nil
	})

	awsCfg, err := awscfg.LoadDefaultConfig(context.Background(),
		awscfg.WithRegion(cfg.S3Region),
		awscfg.WithEndpointResolverWithOptions(resolver),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for MinIO
	})

	return &Service{s3: client, bucket: cfg.S3Bucket, pubURL: cfg.S3PublicURL}, nil
}

// UploadResult holds info about an uploaded file.
type UploadResult struct {
	Key       string
	PublicURL string
	MimeType  string
	SizeBytes int64
	Width     int
	Height    int
}

// Upload validates and stores a file, returning its public URL.
func (s *Service) Upload(ctx context.Context, data []byte, originalName, uploaderID string) (*UploadResult, error) {
	if int64(len(data)) > maxUploadBytes {
		return nil, fmt.Errorf("file too large (max 10MB)")
	}

	// Detect MIME type
	mimeType := mime.TypeByExtension(filepath.Ext(originalName))
	if mimeType == "" {
		// Sniff from content
		if len(data) >= 512 {
			mimeType = sniffMIME(data[:512])
		}
	}
	mimeType = strings.Split(mimeType, ";")[0] // Strip params

	if !AllowedMIME[mimeType] {
		return nil, fmt.Errorf("file type not allowed: %s", mimeType)
	}

	// Generate storage key: uploads/{uploaderID}/{year}/{uuid}.ext
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = mimeToExt(mimeType)
	}
	key := fmt.Sprintf("uploads/%s/%d/%s%s",
		uploaderID,
		time.Now().Year(),
		uuid.New().String(),
		ext,
	)

	// Upload to S3/MinIO
	_, err := s.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(mimeType),
		// Public read — media is served directly
		ACL: "public-read",
	})
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	result := &UploadResult{
		Key:       key,
		PublicURL: fmt.Sprintf("%s/%s", strings.TrimRight(s.pubURL, "/"), key),
		MimeType:  mimeType,
		SizeBytes: int64(len(data)),
	}

	// Get image dimensions if it's an image
	if strings.HasPrefix(mimeType, "image/") && mimeType != "image/svg+xml" {
		if img, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
			result.Width = img.Width
			result.Height = img.Height
		}
	}

	return result, nil
}

// Delete removes a file from storage.
func (s *Service) Delete(ctx context.Context, key string) error {
	_, err := s.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func sniffMIME(data []byte) string {
	// Simple magic byte detection
	signatures := []struct {
		sig  []byte
		mime string
	}{
		{[]byte{0xFF, 0xD8, 0xFF}, "image/jpeg"},
		{[]byte{0x89, 0x50, 0x4E, 0x47}, "image/png"},
		{[]byte{0x47, 0x49, 0x46}, "image/gif"},
		{[]byte{0x52, 0x49, 0x46, 0x46}, "image/webp"},
		{[]byte{0x25, 0x50, 0x44, 0x46}, "application/pdf"},
	}
	for _, sig := range signatures {
		if len(data) >= len(sig.sig) {
			match := true
			for i, b := range sig.sig {
				if data[i] != b {
					match = false
					break
				}
			}
			if match {
				return sig.mime
			}
		}
	}
	return "application/octet-stream"
}

func mimeToExt(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	default:
		return ".bin"
	}
}
