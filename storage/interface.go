package storage

import (
	"context"
	"io"
	"time"
)

// StorageService defines the interface for content storage operations
type StorageService interface {
	// Store saves content data to storage and returns the path
	Store(ctx context.Context, key string, data io.Reader, size int64, contentType string) (string, error)

	// Retrieve gets content data from storage
	Retrieve(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes content data from storage
	Delete(ctx context.Context, path string) error

	// GetURL returns a URL for accessing the content (may be signed/temporary)
	GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)
}
