package storage

import (
	"context"
	"io"
	"time"
	// Assuming your model package path
)

// PresignedURLOptions provides options for generating presigned URLs.
type PresignedURLOptions struct {
	Expiry time.Duration
	// Add other options like content type for upload URLs if needed
}

// StorageService defines the interface for file storage operations.
type StorageService interface {
	Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) (path string, err error)
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	//GetPresignedUploadURL(ctx context.Context, contentID string, fileName string, mimeType string, options PresignedURLOptions) (url *url.URL, additionalHeaders map[string]string, err error)
	GetPresignedDownloadURL(ctx context.Context, path string, options PresignedURLOptions) (url string, err error)
	Delete(ctx context.Context, path string) error
}
