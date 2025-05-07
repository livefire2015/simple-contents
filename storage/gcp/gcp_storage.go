package gcp

import (
	"context"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/torpago/simple-content-service/storage"
)

// GCPStorage implements StorageService using Google Cloud Storage
type GCPStorage struct {
	client     *storage.Client
	bucketName string
}

// NewGCPStorage creates a new GCP storage service
func NewGCPStorage(client *storage.Client, bucketName string) storage.StorageService {
	return &GCPStorage{
		client:     client,
		bucketName: bucketName,
	}
}

// Store saves content data to storage and returns the path
func (s *GCPStorage) Store(ctx context.Context, key string, data io.Reader, size int64, contentType string) (string, error) {
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(key)
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType

	if _, err := io.Copy(writer, data); err != nil {
		writer.Close()
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	return key, nil
}

// Retrieve gets content data from storage
func (s *GCPStorage) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(path)
	return obj.NewReader(ctx)
}

// Delete removes content data from storage
func (s *GCPStorage) Delete(ctx context.Context, path string) error {
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(path)
	return obj.Delete(ctx)
}

// GetURL returns a URL for accessing the content
func (s *GCPStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(path)

	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiry),
	}

	return obj.SignedURL(opts)
}
