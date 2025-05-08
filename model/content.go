package model

import (
	"time"

	"github.com/google/uuid"
)

// Content represents a content item in the system
type Content struct {
	ID          uuid.UUID     `json:"id"`           // Unique identifier (e.g., UUID)
	Status      ContentStatus `json:"status"`       // Processing status
	FileName    string        `json:"file_name"`    // Original name of the file
	MIMEType    string        `json:"mime_type"`    // MIME type of the file
	FileSize    int64         `json:"file_size"`    // Size of the file in bytes
	StoragePath string        `json:"storage_path"` // Path/key in the storage layer
	CreatedBy   string        `json:"created_by"`   // Identifier of the content creator
	CreatedAt   time.Time     `json:"created_at"`   // Timestamp of creation
	UpdatedAt   time.Time     `json:"updated_at"`   // Timestamp of last update
	DeletedAt   *time.Time    `json:"deleted_at,omitempty"`

	// EntityType and EntityID are REMOVED from here
	// as associations are now handled by ContentEntityAssociation.

	Source   string   `json:"source"`             // e.g., "email_attachment", "direct_upload", "slack"
	Metadata Metadata `json:"metadata,omitempty"` // Intrinsic metadata of the content itself
}

// ContentStatus represents the status of a content item.
type ContentStatus string

const (
	StatusCreated  ContentStatus = "created"
	StatusUploaded ContentStatus = "uploaded"
	StatusDone     ContentStatus = "done"
	StatusError    ContentStatus = "error"
	// Add other statuses as needed
)

// Metadata contains additional information about the content
type Metadata map[string]interface{}

// ContentFilter represents filter criteria for content queries
type ContentFilter struct {
	ContentType string
	MinSize     *int64
	MaxSize     *int64
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Metadata    map[string]interface{}
}

// ContentEntityAssociation links a Content item to an external entity
// and can store metadata specific to this particular link.
type ContentEntityAssociation struct {
	ID                  string                 `json:"id"`                   // Unique identifier for the association itself (e.g., UUID)
	ContentID           string                 `json:"content_id"`           // Foreign key to the Content item
	EntityType          string                 `json:"entity_type"`          // Type of the associated entity
	EntityID            string                 `json:"entity_id"`            // ID of the associated entity
	AssociationMetadata map[string]interface{} `json:"association_metadata"` // Metadata specific to this link (e.g., role, version, context)
	CreatedAt           time.Time              `json:"created_at"`           // Timestamp of association creation
	UpdatedAt           time.Time              `json:"updated_at"`           // Timestamp of last update to the association
	CreatedBy           string                 `json:"created_by"`           // Who created this specific association
}

// Example AssociationMetadata:
// For a document linked to an application:
// { "role": "primary_id_proof", "status": "verified" }
// For a template linked to a project:
// { "version_used": "1.2", "customized": true }
