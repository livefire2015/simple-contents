package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/torpago/simple-content-service/model"
)

// ContentRepository defines the interface for content persistence operations
type ContentRepository interface {
	// Create stores a new content item
	Create(ctx context.Context, content *model.Content) error

	// GetByID retrieves a content item by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Content, error)

	// Update updates an existing content item
	Update(ctx context.Context, content *model.Content) error

	// Delete marks a content item as deleted
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves content items based on filter criteria
	List(ctx context.Context, filter model.ContentFilter, offset, limit int) ([]*model.Content, int, error)
}
