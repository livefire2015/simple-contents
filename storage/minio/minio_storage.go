package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/sgao640/simple-contents/storage"
)

// MinioStorage implements StorageService using MinIO
type MinioStorage struct {
	client     *minio.Client
	bucketName string
}

// NewMinioStorage creates a new MinIO storage service
func NewMinioStorage(client *minio.Client, bucketName string) *MinioStorage {
	return &MinioStorage{
		client:     client,
		bucketName: bucketName,
	}
}

// Upload saves content data to storage and returns the path
func (s *MinioStorage) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucketName, key, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return key, nil
}

// Retrieve gets content data from storage
func (s *MinioStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Delete removes content data from storage
func (s *MinioStorage) Delete(ctx context.Context, path string) error {
	return s.client.RemoveObject(ctx, s.bucketName, path, minio.RemoveObjectOptions{})
}

// GetURL returns a URL for accessing the content
func (s *MinioStorage) GetPresignedDownloadURL(ctx context.Context, path string, options storage.PresignedURLOptions) (string, error) {
	// Generate a presigned URL for temporary access
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, path, options.Expiry, nil)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}
