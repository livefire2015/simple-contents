package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sgao640/simple-contents/model"
)

var (
	ErrContentNotFound = errors.New("content not found")
)

// MemoryRepository implements ContentRepository using in-memory storage
type MemoryRepository struct {
	mu       sync.RWMutex
	contents map[uuid.UUID]*model.Content
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		contents: make(map[uuid.UUID]*model.Content),
	}
}

// Create stores a new content item
func (r *MemoryRepository) CreateContent(ctx context.Context, content *model.Content) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if content.ID == uuid.Nil {
		content.ID = uuid.New()
	}

	now := time.Now()
	content.CreatedAt = now
	content.UpdatedAt = now

	r.contents[content.ID] = content
	return nil
}

// GetByID retrieves a content item by its ID
func (r *MemoryRepository) GetContentByID(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	content, exists := r.contents[id]
	if !exists || content.DeletedAt != nil {
		return nil, ErrContentNotFound
	}

	// Return a copy to prevent modification of the stored data
	contentCopy := *content
	return &contentCopy, nil
}

// Update updates an existing content item
func (r *MemoryRepository) UpdateContent(ctx context.Context, content *model.Content) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.contents[content.ID]
	if !exists || existing.DeletedAt != nil {
		return ErrContentNotFound
	}

	content.CreatedAt = existing.CreatedAt
	content.UpdatedAt = time.Now()

	r.contents[content.ID] = content
	return nil
}

// Delete marks a content item as deleted
func (r *MemoryRepository) DeleteContent(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	content, exists := r.contents[id]
	if !exists || content.DeletedAt != nil {
		return ErrContentNotFound
	}

	now := time.Now()
	content.DeletedAt = &now
	return nil
}

// List retrieves content items based on filter criteria
func (r *MemoryRepository) ListContent(ctx context.Context, filter model.ContentFilter, offset, limit int) ([]*model.Content, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filteredContents []*model.Content

	// Apply filters
	for _, content := range r.contents {
		if content.DeletedAt != nil {
			continue
		}

		if filter.MIMEType != "" && content.MIMEType != filter.MIMEType {
			continue
		}

		if filter.MinSize != nil && content.FileSize < *filter.MinSize {
			continue
		}

		if filter.MaxSize != nil && content.FileSize > *filter.MaxSize {
			continue
		}

		if filter.CreatedFrom != nil && content.CreatedAt.Before(*filter.CreatedFrom) {
			continue
		}

		if filter.CreatedTo != nil && content.CreatedAt.After(*filter.CreatedTo) {
			continue
		}

		// Check metadata filters if any
		if len(filter.Metadata) > 0 {
			match := true
			for k, v := range filter.Metadata {
				if contentValue, exists := content.Metadata[k]; !exists || contentValue != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		// Create a copy to prevent modification of the stored data
		contentCopy := *content
		filteredContents = append(filteredContents, &contentCopy)
	}

	// Calculate total count
	totalCount := len(filteredContents)

	// Apply pagination
	if offset >= len(filteredContents) {
		return []*model.Content{}, totalCount, nil
	}

	end := offset + limit
	if end > len(filteredContents) {
		end = len(filteredContents)
	}

	return filteredContents[offset:end], totalCount, nil
}
