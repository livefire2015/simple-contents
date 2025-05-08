package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/torpago/simple-content-service/repository"
	"github.com/torpago/simple-content-service/storage"
	"github.com/torpago/simple-contents/model"
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
	FileName  string
	MIMEType  string
	FileSize  int64
	CreatedBy string
	// ** Crucial for association **
	EntityType string // e.g., common.EntityTypeTransaction
	EntityID   string // e.g., the specific transaction ID
	// ** End crucial for association **
	Source   string
	Metadata model.Metadata
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

// AssociateContentInput defines the input for associating content with an entity
type AssociateContentInput struct {
	ContentID           string                 `json:"content_id"`
	EntityType          string                 `json:"entity_type"`
	EntityID            string                 `json:"entity_id"`
	AssociationMetadata map[string]interface{} `json:"association_metadata"`
	AssociatedBy        string                 `json:"associated_by"` // User/service performing the association
}

// AssociateContent links an existing content item to an entity.
func (s *ContentService) AssociateContent(ctx context.Context, input AssociateContentInput) (*model.ContentEntityAssociation, error) {
	// 1. Validate that the content item exists
	_, err := s.repo.GetContentByID(ctx, input.ContentID)
	if err != nil {
		return nil, fmt.Errorf("content with ID %s not found: %w", input.ContentID, err)
	}

	// 2. (Optional) Validate that the associating entity exists by calling another service or based on known types.
	// This depends on your system's architecture.

	// 3. Check for existing association if you don't want duplicates (based on unique constraint)
	existingAssoc, err := s.repo.GetAssociationByLink(ctx, input.ContentID, input.EntityType, input.EntityID)
	if err != nil && err != repository.ErrNotFound { // Assuming ErrNotFound is a distinct error type
		return nil, fmt.Errorf("error checking for existing association: %w", err)
	}
	if existingAssoc != nil {
		// You might want to update the existing one or return an error, based on policy
		return nil, fmt.Errorf("content %s is already associated with entity %s/%s (association ID: %s)",
			input.ContentID, input.EntityType, input.EntityID, existingAssoc.ID)
	}

	association := &model.ContentEntityAssociation{
		ID:                  uuid.NewString(), // Generate new ID for the association
		ContentID:           input.ContentID,
		EntityType:          input.EntityType,
		EntityID:            input.EntityID,
		AssociationMetadata: input.AssociationMetadata,
		CreatedBy:           input.AssociatedBy,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	if err := s.repo.CreateAssociation(ctx, association); err != nil {
		return nil, fmt.Errorf("failed to create association: %w", err)
	}

	return association, nil
}

// GetContentForEntity retrieves content items linked to a specific entity.
func (s *ContentService) GetContentForEntity(ctx context.Context, entityType string, entityID string, options repository.ListOptions) ([]*model.Content, int64, error) {
	if entityType == "" || entityID == "" {
		return nil, 0, fmt.Errorf("entityType and entityID are required")
	}
	// This service method now calls the repository method that handles the join
	return s.repo.ListContentByEntity(ctx, entityType, entityID, options)
}

// Inside your service.ContentService

// Assume s.storage has a method StatObject(ctx, storagePath) (string, error) that returns ObjectMetadata
// type ObjectMetadata struct {
//    Size int64
//    ContentType string // Could also get the ContentType set by the storage service
//    // Other relevant metadata
// }

func (s *ContentService) MarkContentAsUploaded(ctx context.Context, contentID string, storagePath string) (*model.Content, error) {
	content, err := s.repo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("content with ID %s not found: %w", contentID, err)
	}

	if content.Status != model.StatusCreated && content.Status != model.StatusError {
		return nil, fmt.Errorf("cannot mark content as uploaded, current status: %s", content.Status)
	}

	// At this point, the file is already in the storage (e.g., S3, Minio)
	// We need to fetch the first 512 bytes from storage to verify MIME type.

	var detectedMIMEType string
	var actualFileSize int64 // Get this from storage if possible

	// --- Hypothetical steps to get header from storage ---
	// This part is pseudo-code as it depends on your StorageService interface and implementation
	// You might need to add a method like `GetFirstNBytes(path, n)` to your StorageService
	// or use the Download method with a ranged request if supported.

	fileHeaderReader, err := s.storage.DownloadRange(ctx, clientProvidedStoragePath, 0, 511) // Hypothetical method
	if err != nil {
		// Handle error: maybe can't access file, or file too small
		// You might decide to trust client MIME or mark as error
		fmt.Printf("Warning: could not download file header for MIME detection from storage for %s: %v\n", contentID, err)
		// Fallback or error out based on policy
		detectedMIMEType = content.MIMEType // Or mark as error/unknown
	} else {
		defer fileHeaderReader.Close()
		headerBytes, readErr := io.ReadAll(fileHeaderReader)
		if readErr != nil {
			fmt.Printf("Warning: could not read file header for MIME detection for %s: %v\n", contentID, readErr)
			detectedMIMEType = content.MIMEType // Fallback
		} else {
			detectedMIMEType = http.DetectContentType(headerBytes)
			fmt.Printf("Post-upload MIME check for %s. Client: %s, Server detected: %s\n", contentID, content.MIMEType, detectedMIMEType)
			if content.MIMEType != detectedMIMEType {
				// Your policy here: update, log, reject, etc.
				content.MIMEType = detectedMIMEType // Example: update to server-detected
			}
		}
	}
	// --- End hypothetical steps ---

	// 1. Get metadata from storage service
	objectMetadata, err := s.storage.StatObject(ctx, storagePath) // This is a hypothetical method you'd add to your StorageService interface and implement
	if err != nil {
		// Potentially mark content as error, or retry, or log and proceed with client-provided size if that's your policy
		s.repo.UpdateStatus(ctx, contentID, model.StatusError)
		return nil, fmt.Errorf("failed to get object metadata from storage for %s: %w", storagePath, err)
	}

	actualFileSize := objectMetadata.Size
	// You could also trust the ContentType from storage if it's reliable,
	// or perform your own header download + DetectContentType as discussed before.
	// detectedMIMETypeFromStorage := objectMetadata.ContentType

	// (Optional: MIME Type detection by downloading the first 512 bytes, as discussed previously)
	// ... your MIME detection logic here if you don't trust storage-provided MIME ...
	// verifiedMIMEType := ...

	content.StoragePath = storagePath
	content.FileSize = actualFileSize // Use the authoritative size from storage
	content.Status = model.StatusUploaded
	// content.MIMEType = verifiedMIMEType // Update if you re-verified
	content.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, content); err != nil {
		// Consider what to do if DB update fails. File is in storage.
		// Maybe a retry mechanism or an "undo" by deleting from storage is too risky / complex here.
		// Logging this inconsistency is critical.
		return nil, fmt.Errorf("failed to update content record after upload confirmation: %w", err)
	}

	return content, nil
}
