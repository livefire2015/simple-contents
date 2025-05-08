package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sgao640/simple-contents/model"
)

var (
	ErrContentNotFound = errors.New("content not found")
)

// PostgresRepository implements ContentRepository using PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// contentDB is a database model for content
type contentDB struct {
	ID          uuid.UUID      `db:"id"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	MIMEType    string         `db:"mime_type"`
	FileSize    int64          `db:"file_size"`
	Path        string         `db:"path"`
	Metadata    sql.NullString `db:"metadata"` // JSON stored as string
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
}

// toModel converts a database model to a domain model
func (c *contentDB) toModel() (*model.Content, error) {
	content := &model.Content{
		ID:          c.ID,
		FileName:    c.Name,
		MIMEType:    c.MIMEType,
		FileSize:    c.FileSize,
		StoragePath: c.Path,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.DeletedAt.Valid {
		content.DeletedAt = &c.DeletedAt.Time
	}

	// Parse metadata JSON
	if c.Metadata.Valid {
		var metadata model.Metadata
		if err := json.Unmarshal([]byte(c.Metadata.String), &metadata); err != nil {
			return nil, err
		}
		content.Metadata = metadata
	} else {
		content.Metadata = make(model.Metadata)
	}

	return content, nil
}

// fromModel converts a domain model to a database model
func fromModel(content *model.Content) (*contentDB, error) {
	dbContent := &contentDB{
		ID:        content.ID,
		Name:      content.FileName,
		MIMEType:  content.MIMEType,
		FileSize:  content.FileSize,
		Path:      content.StoragePath,
		CreatedAt: content.CreatedAt,
		UpdatedAt: content.UpdatedAt,
	}

	if content.DeletedAt != nil {
		dbContent.DeletedAt = sql.NullTime{
			Time:  *content.DeletedAt,
			Valid: true,
		}
	}

	// Convert metadata to JSON
	if len(content.Metadata) > 0 {
		metadataBytes, err := json.Marshal(content.Metadata)
		if err != nil {
			return nil, err
		}
		dbContent.Metadata = sql.NullString{
			String: string(metadataBytes),
			Valid:  true,
		}
	}

	return dbContent, nil
}

// Create stores a new content item
func (r *PostgresRepository) CreateContent(ctx context.Context, content *model.Content) error {
	if content.ID == uuid.Nil {
		content.ID = uuid.New()
	}

	now := time.Now()
	content.CreatedAt = now
	content.UpdatedAt = now

	dbContent, err := fromModel(content)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO contents (
			id, name, description, content_type, size, path, metadata, created_at, updated_at
		) VALUES (
			:id, :name, :description, :content_type, :size, :path, :metadata, :created_at, :updated_at
		)
	`

	_, err = r.db.NamedExecContext(ctx, query, dbContent)
	return err
}

// GetByID retrieves a content item by its ID
func (r *PostgresRepository) GetContentByID(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	query := `
		SELECT * FROM contents 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var dbContent contentDB
	if err := r.db.GetContext(ctx, &dbContent, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrContentNotFound
		}
		return nil, err
	}

	return dbContent.toModel()
}

// Update updates an existing content item
func (r *PostgresRepository) UpdateContent(ctx context.Context, content *model.Content) error {
	content.UpdatedAt = time.Now()

	dbContent, err := fromModel(content)
	if err != nil {
		return err
	}

	query := `
		UPDATE contents SET
			name = :name,
			description = :description,
			content_type = :content_type,
			size = :size,
			path = :path,
			metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	result, err := r.db.NamedExecContext(ctx, query, dbContent)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrContentNotFound
	}

	return nil
}

// Delete marks a content item as deleted
func (r *PostgresRepository) DeleteContent(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE contents SET
			deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrContentNotFound
	}

	return nil
}

// buildWhereClause constructs the WHERE clause for filtering
func buildWhereClause(filter model.ContentFilter) (string, []interface{}) {
	where := "deleted_at IS NULL"
	var params []interface{}
	paramCount := 1

	if filter.ContentType != "" {
		where += " AND content_type = $" + string(paramCount)
		params = append(params, filter.ContentType)
		paramCount++
	}

	if filter.MinSize != nil {
		where += " AND size >= $" + string(paramCount)
		params = append(params, *filter.MinSize)
		paramCount++
	}

	if filter.MaxSize != nil {
		where += " AND size <= $" + string(paramCount)
		params = append(params, *filter.MaxSize)
		paramCount++
	}

	if filter.CreatedFrom != nil {
		where += " AND created_at >= $" + string(paramCount)
		params = append(params, *filter.CreatedFrom)
		paramCount++
	}

	if filter.CreatedTo != nil {
		where += " AND created_at <= $" + string(paramCount)
		params = append(params, *filter.CreatedTo)
		paramCount++
	}

	// Metadata filtering is more complex with JSON
	if len(filter.Metadata) > 0 {
		for key, value := range filter.Metadata {
			where += " AND metadata->$" + string(paramCount) + " = $" + string(paramCount+1)
			params = append(params, key, value)
			paramCount += 2
		}
	}

	return where, params
}

// List retrieves content items based on filter criteria
func (r *PostgresRepository) ListContent(ctx context.Context, filter model.ContentFilter, offset, limit int) ([]*model.Content, int, error) {
	whereClause, params := buildWhereClause(filter)

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM contents WHERE " + whereClause
	var totalCount int
	if err := r.db.GetContext(ctx, &totalCount, countQuery, params...); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := "SELECT * FROM contents WHERE " + whereClause + " ORDER BY created_at DESC LIMIT $" + string(len(params)+1) + " OFFSET $" + string(len(params)+2)
	params = append(params, limit, offset)

	var dbContents []contentDB
	if err := r.db.SelectContext(ctx, &dbContents, query, params...); err != nil {
		return nil, 0, err
	}

	// Convert to domain models
	contents := make([]*model.Content, len(dbContents))
	for i, dbContent := range dbContents {
		content, err := dbContent.toModel()
		if err != nil {
			return nil, 0, err
		}
		contents[i] = content
	}

	return contents, totalCount, nil
}
