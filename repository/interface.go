package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/livefire2015/simple-contents/model" // Adjust import path as needed
)

// ListOptions remains the same (for pagination, sorting)
type ListOptions struct {
	Page        int
	PageSize    int
	SortBy      string
	ReturnTotal bool // Whether to calculate and return total count
}

// ContentRepository defines the interface for content and association persistence.
type ContentRepository interface {
	// --- Content Specific Methods ---
	CreateContent(ctx context.Context, content *model.Content) error
	GetContentByID(ctx context.Context, id uuid.UUID) (*model.Content, error)
	ListContent(ctx context.Context, filter model.ContentFilter, offset int, limit int) ([]*model.Content, int, error)
	UpdateContent(ctx context.Context, content *model.Content) error // For metadata, status, etc.
	DeleteContent(ctx context.Context, id uuid.UUID) error           // This would cascade to associations if DB constraints are set

	// // --- Association Specific Methods ---
	// CreateAssociation(ctx context.Context, association *model.ContentEntityAssociation) error
	// GetAssociationByID(ctx context.Context, associationID string) (*model.ContentEntityAssociation, error)
	// // Get a specific association if its ID isn't known but the linked items are.
	// GetAssociationByLink(ctx context.Context, contentID, entityType, entityID string) (*model.ContentEntityAssociation, error)
	// UpdateAssociation(ctx context.Context, association *model.ContentEntityAssociation) error // e.g., to update metadata or re-link (less common)
	// DeleteAssociation(ctx context.Context, associationID string) error
	// // Alternative: DeleteAssociationByLink(ctx context.Context, contentID, entityType, entityID string) error

	// // --- Querying Methods (involving associations) ---

	// // List content associated with a specific entity.
	// // The implementation will join `contents` with `content_entity_associations`.
	// ListContentByEntity(ctx context.Context, entityType string, entityID string, options ListOptions) (contents []*model.Content, total int64, err error)

	// // List associations for a given entity (useful if you want the association metadata too).
	// ListAssociationsByEntity(ctx context.Context, entityType string, entityID string, options ListOptions) (associations []*model.ContentEntityAssociation, total int64, err error)

	// // List entities (via associations) linked to a specific content item.
	// ListAssociationsByContent(ctx context.Context, contentID string, options ListOptions) (associations []*model.ContentEntityAssociation, total int64, err error)

	// (Optional) Search content based on association metadata (more complex query)
	// SearchContentByAssociationMetadata(ctx context.Context, entityType string, entityID string, metadataQuery map[string]interface{}, options ListOptions) ([]*model.Content, int64, error)
}

var ErrContentNotFound = errors.New("content not found")
