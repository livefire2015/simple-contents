package model

import (
	"time"

	"github.com/google/uuid"
)

// Content represents a content item in the system
type Content struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ContentType string     `json:"content_type"`
	Size        int64      `json:"size"`
	Path        string     `json:"path"`
	Metadata    Metadata   `json:"metadata"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

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
