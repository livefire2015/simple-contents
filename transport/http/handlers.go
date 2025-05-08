package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/sgao640/simple-contents/model"
	"github.com/sgao640/simple-contents/service"
)

// ContentHandler handles HTTP requests for content operations
type ContentHandler struct {
	contentService *service.ContentService
}

// NewContentHandler creates a new content HTTP handler
func NewContentHandler(contentService *service.ContentService) *ContentHandler {
	return &ContentHandler{
		contentService: contentService,
	}
}

// RegisterRoutes registers HTTP routes for content operations
func (h *ContentHandler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1/contents", func(r chi.Router) {
		r.Post("/", h.CreateContent)
		r.Get("/", h.ListContents)
		r.Get("/{id}", h.GetContent)
		r.Put("/{id}", h.UpdateContent)
		r.Delete("/{id}", h.DeleteContent)
		r.Get("/{id}/data", h.GetContentData)
		r.Get("/{id}/url", h.GetContentURL)
	})
}

// errorResponse sends an error response with the given status code and message
func errorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// CreateContent handles the creation of new content
func (h *ContentHandler) CreateContent(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}

	// Get form values
	name := r.FormValue("name")
	metadataStr := r.FormValue("metadata")

	// Parse metadata if provided
	var metadata model.Metadata
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid metadata format")
			return
		}
	} else {
		metadata = make(model.Metadata)
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Create content
	input := service.CreateContentInput{
		FileName: name,
		MIMEType: header.Header.Get("Content-Type"),
		FileSize: header.Size,
		Metadata: metadata,
	}

	content, err := h.contentService.CreateContent(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			errorResponse(w, http.StatusBadRequest, err.Error())
		} else {
			errorResponse(w, http.StatusInternalServerError, "Failed to create content")
		}
		return
	}

	// Return created content
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(content)
}

// GetContent handles retrieving content metadata by ID
func (h *ContentHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid content ID")
		return
	}

	content, err := h.contentService.GetContent(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrContentNotFound) {
			errorResponse(w, http.StatusNotFound, "Content not found")
		} else {
			errorResponse(w, http.StatusInternalServerError, "Failed to retrieve content")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

// UpdateContent handles updating content metadata
func (h *ContentHandler) UpdateContent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid content ID")
		return
	}

	var input struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Metadata    model.Metadata `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updateInput := service.UpdateContentInput{
		ID:       id,
		FileName: input.Name,
		Metadata: input.Metadata,
	}

	content, err := h.contentService.UpdateContent(r.Context(), updateInput)
	if err != nil {
		if errors.Is(err, service.ErrContentNotFound) {
			errorResponse(w, http.StatusNotFound, "Content not found")
		} else if errors.Is(err, service.ErrInvalidInput) {
			errorResponse(w, http.StatusBadRequest, err.Error())
		} else {
			errorResponse(w, http.StatusInternalServerError, "Failed to update content")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

// DeleteContent handles deleting content
func (h *ContentHandler) DeleteContent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid content ID")
		return
	}

	err = h.contentService.DeleteContent(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrContentNotFound) {
			errorResponse(w, http.StatusNotFound, "Content not found")
		} else {
			errorResponse(w, http.StatusInternalServerError, "Failed to delete content")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetContentData handles retrieving content data
func (h *ContentHandler) GetContentData(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid content ID")
		return
	}

	data, content, err := h.contentService.GetContentData(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrContentNotFound) {
			errorResponse(w, http.StatusNotFound, "Content not found")
		} else {
			errorResponse(w, http.StatusInternalServerError, "Failed to retrieve content data")
		}
		return
	}
	defer data.Close()

	// Set appropriate headers
	w.Header().Set("Content-Type", content.MIMEType)
	w.Header().Set("Content-Disposition", "attachment; filename="+content.FileName)
	w.Header().Set("Content-Length", strconv.FormatInt(content.FileSize, 10))

	// Stream the data to the response
	_, err = io.Copy(w, data)
	if err != nil {
		// Log the error but don't return a response as headers have already been sent
		// log.Printf("Error streaming content data: %v", err)
	}
}

// GetContentURL handles generating a URL for accessing content
func (h *ContentHandler) GetContentURL(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid content ID")
		return
	}

	// Parse expiry time from query parameter (default to 1 hour)
	expiryStr := r.URL.Query().Get("expiry")
	expiry := 1 * time.Hour
	if expiryStr != "" {
		expirySeconds, err := strconv.ParseInt(expiryStr, 10, 64)
		if err == nil && expirySeconds > 0 {
			expiry = time.Duration(expirySeconds) * time.Second
		}
	}

	url, err := h.contentService.GetContentURL(r.Context(), id, expiry)
	if err != nil {
		if errors.Is(err, service.ErrContentNotFound) {
			errorResponse(w, http.StatusNotFound, "Content not found")
		} else {
			errorResponse(w, http.StatusInternalServerError, "Failed to generate content URL")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

// ListContents handles listing content items
func (h *ContentHandler) ListContents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	// Parse pagination parameters
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))

	// Parse filter parameters
	contentType := query.Get("contentType")

	var minSize, maxSize *int64
	if minSizeStr := query.Get("minSize"); minSizeStr != "" {
		if val, err := strconv.ParseInt(minSizeStr, 10, 64); err == nil {
			minSize = &val
		}
	}
	if maxSizeStr := query.Get("maxSize"); maxSizeStr != "" {
		if val, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			maxSize = &val
		}
	}

	var createdFrom, createdTo *time.Time
	if fromStr := query.Get("createdFrom"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			createdFrom = &t
		}
	}
	if toStr := query.Get("createdTo"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			createdTo = &t
		}
	}

	// Parse metadata filter
	var metadata map[string]interface{}
	if metadataStr := query.Get("metadata"); metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid metadata format")
			return
		}
	}

	input := service.ListContentInput{
		MIMEType:    contentType,
		MinSize:     minSize,
		MaxSize:     maxSize,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		Metadata:    metadata,
		Page:        page,
		PageSize:    pageSize,
	}

	result, err := h.contentService.ListContent(r.Context(), input)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to list content")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
