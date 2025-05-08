package s3

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sgao640/simple-contents/storage"
)

// S3Storage implements StorageService using AWS S3
type S3Storage struct {
	client     *s3.Client
	bucketName string
	region     string
}

// NewS3Storage creates a new S3 storage service
func NewS3Storage(client *s3.Client, bucketName, region string) storage.StorageService {
	return &S3Storage{
		client:     client,
		bucketName: bucketName,
		region:     region,
	}
}

// Store saves content data to storage and returns the path
func (s *S3Storage) Store(ctx context.Context, key string, data io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	return key, nil
}

// Retrieve gets content data from storage
func (s *S3Storage) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

// Delete removes content data from storage
func (s *S3Storage) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	return err
}

// GetURL returns a URL for accessing the content
func (s *S3Storage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}
