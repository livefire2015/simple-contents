package service

import (
	"context"
	"errors"
	"io"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/torpago/simple-content-service/model"
	"github.com/torpago/simple-content-service/repository"
	"github.com/torpago/simple-content-service/storage"
)

var (
	ErrContentNotFound = errors.New("content not found")
	ErrInvalidInput    = errors.New("invalid input parameters")
)

// ContentService handles business logic for content operations
type ContentService struct {
	repo    repository.ContentRepository
	storage storage.StorageService
}

// NewContentService creates a new content service
func NewContentService(repo repository.ContentRepository, storage storage.StorageService) *ContentService {
	return &ContentService{
		repo:    repo,
		storage: storage,
	}
}

// CreateContentInput represents input for creating content
type CreateContentInput struct {
	Name        string
	Description string
	ContentType string
	Data        io.Reader
	Size        int64
	Metadata    model.Metadata
}

// CreateContent creates a new content item
func (s *ContentService) CreateContent(ctx context.Context, input CreateContentInput) (*model.Content, error) {
	if input.Name == "" || input.ContentType == "" || input.Data == nil || input.Size <= 0 {
		return nil, ErrInvalidInput
	}

	// Generate a unique ID for the content
	contentID := uuid.New()

	// Create a storage key based on content ID and name
	storageKey := path.Join(contentID.String(), input.Name)

	// Store the content data
	storagePath, err := s.storage.Store(ctx, storageKey, input.Data, input.Size, input.ContentType)
	if err != nil {
		return nil, err
	}

	// Create the content record
	content := &model.Content{
		ID:          contentID,
		Name:        input.Name,
		Description: input.Description,
		ContentType: input.ContentType,
		Size:        input.Size,
		Path:        storagePath,
		Metadata:    input.Metadata,
	}

	if err := s.repo.Create(ctx, content); err != nil {
		// Clean up storage if repository creation fails
		_ = s.storage.Delete(ctx, storagePath)
		return nil, err
	}

	return content, nil
}

// GetContent retrieves a content item by ID
func (s *ContentService) GetContent(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	content, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrContentNotFound) {
			return nil, ErrContentNotFound
		}
		return nil, err
	}

	return content, nil
}

// GetContentData retrieves the data for a content item
func (s *ContentService) GetContentData(ctx context.Context, id uuid.UUID) (io.ReadCloser, *model.Content, error) {
	content, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrContentNotFound) {
			return nil, nil, ErrContentNotFound
		}
		return nil, nil, err
	}

	data, err := s.storage.Retrieve(ctx, content.Path)
	if err != nil {
		return nil, nil, err
	}

	return data, content, nil
}

// UpdateContentInput represents input for updating content
type UpdateContentInput struct {
	ID          uuid.UUID
	Name        string
	Description string
	Metadata    model.Metadata
}

// UpdateContent updates a content item
func (s *ContentService) UpdateContent(ctx context.Context, input UpdateContentInput) (*model.Content, error) {
	if input.ID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	content, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, repository.ErrContentNotFound) {
			return nil, ErrContentNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if input.Name != "" {
		content.Name = input.Name
	}
	if input.Description != "" {
		content.Description = input.Description
	}
	if input.Metadata != nil {
		content.Metadata = input.Metadata
	}

	if err := s.repo.Update(ctx, content); err != nil {
		return nil, err
	}

	return content, nil
}

// DeleteContent deletes a content item
func (s *ContentService) DeleteContent(ctx context.Context, id uuid.UUID) error {
	content, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrContentNotFound) {
			return ErrContentNotFound
		}
		return err
	}

	// Delete from repository first
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Then delete from storage
	// Note: We don't return storage deletion errors to the caller
	// as the content is already marked as deleted in the repository
	_ = s.storage.Delete(ctx, content.Path)

	return nil
}

// ListContentInput represents input for listing content
type ListContentInput struct {
	ContentType string
	MinSize     *int64
	MaxSize     *int64
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Metadata    map[string]interface{}
	Page        int
	PageSize    int
}

// ListContentResult represents the result of listing content
type ListContentResult struct {
	Items      []*model.Content
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// ListContent lists content items based on filter criteria
func (s *ContentService) ListContent(ctx context.Context, input ListContentInput) (*ListContentResult, error) {
	// Set default pagination values if not provided
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	// Calculate offset for pagination
	offset := (input.Page - 1) * input.PageSize

	// Create filter from input
	filter := model.ContentFilter{
		ContentType: input.ContentType,
		MinSize:     input.MinSize,
		MaxSize:     input.MaxSize,
		CreatedFrom: input.CreatedFrom,
		CreatedTo:   input.CreatedTo,
		Metadata:    input.Metadata,
	}

	// Get content items
	items, totalCount, err := s.repo.List(ctx, filter, offset, input.PageSize)
	if err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := totalCount / input.PageSize
	if totalCount%input.PageSize > 0 {
		totalPages++
	}

	return &ListContentResult{
		Items:      items,
		TotalCount: totalCount,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetContentURL generates a URL for accessing content
func (s *ContentService) GetContentURL(ctx context.Context, id uuid.UUID, expiry time.Duration) (string, error) {
	content, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrContentNotFound) {
			return "", ErrContentNotFound
		}
		return "", err
	}

	return s.storage.GetURL(ctx, content.Path, expiry)
}
