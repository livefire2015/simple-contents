package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/sgao640/simple-contents/model" // Assuming your model package path
)

// PresignedURLOptions provides options for generating presigned URLs.
type PresignedURLOptions struct {
	Expiry time.Duration
	// Add other options like content type for upload URLs if needed
}

// StorageService defines the interface for file storage operations.
type StorageService interface {
	Upload(ctx context.Context, content *model.Content, reader io.Reader) (storagePath string, err error)
	Download(ctx context.Context, storagePath string) (io.ReadCloser, error)
	GetPresignedUploadURL(ctx context.Context, contentID string, fileName string, mimeType string, options PresignedURLOptions) (url *url.URL, additionalHeaders map[string]string, err error)
	GetPresignedDownloadURL(ctx context.Context, storagePath string, fileName string, options PresignedURLOptions) (url *url.URL, err error)
	Delete(ctx context.Context, storagePath string) error
}
