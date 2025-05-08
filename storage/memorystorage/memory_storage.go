package memorystorage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/livefire2015/simple-contents/storage"
)

var (
	ErrContentNotFound = errors.New("content not found in storage")
)

// MemoryStorage implements StorageService using in-memory storage
type MemoryStorage struct {
	mu      sync.RWMutex
	storage map[string][]byte
}

// NewMemoryStorage creates a new in-memory storage service
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		storage: make(map[string][]byte),
	}
}

// Store saves content data to storage and returns the path
func (s *MemoryStorage) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read all data from the reader
	content, err := io.ReadAll(data)
	if err != nil {
		return "", err
	}

	// Store the data with the key as the path
	s.storage[key] = content

	return key, nil
}

// Retrieve gets content data from storage
func (s *MemoryStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	content, exists := s.storage[path]
	if !exists {
		return nil, ErrContentNotFound
	}

	// Return a reader for the content
	return io.NopCloser(bytes.NewReader(content)), nil
}

// Delete removes content data from storage
func (s *MemoryStorage) Delete(ctx context.Context, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.storage[path]; !exists {
		return ErrContentNotFound
	}

	delete(s.storage, path)
	return nil
}

// GetURL returns a URL for accessing the content
// For in-memory storage, this is just a placeholder as there's no real URL
func (s *MemoryStorage) GetPresignedDownloadURL(ctx context.Context, path string, options storage.PresignedURLOptions) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.storage[path]; !exists {
		return "", ErrContentNotFound
	}

	// For in-memory storage, we just return a fake URL
	return "memory://" + path, nil
}
