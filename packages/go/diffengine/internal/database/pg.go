package db

import (
	"database/sql"
	"fmt"

	"github.com/specterops/bloodhound/packages/go/diffengine/internal/diff"
	"github.com/specterops/bloodhound/packages/go/diffengine/internal/models"
	"gorm.io/gorm"
)

type DatabaseService struct {
	db *sql.DB
}

type DatabaseOperations interface {
	Close() error
	FetchSchemaByName(name string) (*models.Schema, error)
	CreateSchema(schema *models.Schema) error
	UpdateSchema(schema *models.Schema, diff *diff.DiffResult) error
	DeleteSchema(name string) error
}

func NewDatabaseService(db *gorm.DB) (DatabaseOperations, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error retrieving sql database: %v", err)
	}

	return &DatabaseService{
		db: sqlDB,
	}, nil
}

func (s *DatabaseService) Close() error {
	return s.db.Close()
}

// FetchSchemaByName retrieves a schema and all its related data using raw SQL
func (s *DatabaseService) FetchSchemaByName(name string) (*models.Schema, error) {
	var schema models.Schema

	// Fetch main schema record
	err := s.db.QueryRow(`
		SELECT id, name, version, created_at, updated_at
		FROM schemas WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(&schema.ID, &schema.Name, &schema.Version, &schema.CreatedAt, &schema.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Schema doesn't exist
	}
	if err != nil {
		return nil, err
	}

	// Fetch fields
	rows, err := s.db.Query(`
		SELECT id, schema_id, field_name, field_type, is_required
		FROM schema_fields WHERE schema_id = $1 AND deleted_at IS NULL
	`, schema.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var field models.Field
		if err := rows.Scan(&field.ID, &field.SchemaID, &field.Name, &field.Type, &field.IsRequired); err != nil {
			return nil, err
		}
		schema.Fields = append(schema.Fields, field)
	}

	// Fetch metadata
	var metadata models.Metadata
	err = s.db.QueryRow(`
		SELECT id, schema_id, description, author
		FROM schema_metadata WHERE schema_id = $1 AND deleted_at IS NULL
	`, schema.ID).Scan(&metadata.ID, &metadata.SchemaID, &metadata.Description, &metadata.Author)

	if err == nil {
		schema.Metadata = &metadata
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	// Fetch tags
	tagRows, err := s.db.Query(`
		SELECT id, tag_key, tag_value FROM schema_tags WHERE schema_id = $1 AND deleted_at IS NULL
	`, schema.ID)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var tag models.Tag
		if err := tagRows.Scan(&tag.ID, &tag.Key, &tag.Value); err != nil {
			return nil, err
		}
		schema.Tags = append(schema.Tags, tag)
	}
	return &schema, nil
}

// CreateSchema inserts a new schema with all related data using raw SQL
func (s *DatabaseService) CreateSchema(schema *models.Schema) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert main schema
	var schemaID uint
	err = tx.QueryRow(`
		INSERT INTO schemas (name, version, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, schema.Name, schema.Version).Scan(&schemaID)
	if err != nil {
		return err
	}

	// Insert fields
	for _, field := range schema.Fields {
		_, err = tx.Exec(`
			INSERT INTO schema_fields (schema_id, field_name, field_type, is_required, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`, schemaID, field.Name, field.Type, field.IsRequired)
		if err != nil {
			return err
		}
	}

	// Insert metadata
	if schema.Metadata != nil {
		_, err = tx.Exec(`
			INSERT INTO schema_metadata (schema_id, description, author, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
		`, schemaID, schema.Metadata.Description, schema.Metadata.Author)
		if err != nil {
			return err
		}
	}

	// Insert tags
	for _, tag := range schema.Tags {
		_, err = tx.Exec(`
		INSERT INTO schema_tags (schema_id, tag_key, tag_value, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		`, schemaID, tag.Key, tag.Value)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateSchema applies diff changes to existing schema using raw SQL
func (s *DatabaseService) UpdateSchema(schema *models.Schema, diff *diff.DiffResult) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update schema version and timestamp
	_, err = tx.Exec(`
		UPDATE schemas
		SET version = $1, updated_at = NOW()
		WHERE id = $2
	`, schema.Version, schema.ID)
	if err != nil {
		return err
	}

	// Delete removed fields (using natural key)
	for _, field := range diff.FieldsToDelete {
		_, err = tx.Exec(`
			DELETE FROM schema_fields
			WHERE schema_id = $1 AND field_name = $2
		`, schema.ID, field.Name)
		if err != nil {
			return err
		}
	}

	// Update existing fields
	for _, fieldUpdate := range diff.FieldsToUpdate {
		_, err = tx.Exec(`
			UPDATE schema_fields
			SET field_type = $1, is_required = $2, updated_at = NOW()
			WHERE schema_id = $3 AND field_name = $4
		`, fieldUpdate.NewType, fieldUpdate.NewRequired, schema.ID, fieldUpdate.Name)
		if err != nil {
			return err
		}
	}

	// Insert new fields
	for _, field := range diff.FieldsToAdd {
		_, err = tx.Exec(`
			INSERT INTO schema_fields (schema_id, field_name, field_type, is_required, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`, schema.ID, field.Name, field.Type, field.IsRequired)
		if err != nil {
			return err
		}
	}

	// Handle metadata
	if diff.MetadataChange != nil {
		switch diff.MetadataChange.Action {
		case "created", "updated":
			if schema.Metadata != nil {
				// Try update first
				result, err := tx.Exec(`
					UPDATE schema_metadata
					SET description = $1, author = $2, updated_at = NOW()
					WHERE schema_id = $3
				`, schema.Metadata.Description, schema.Metadata.Author, schema.ID)
				if err != nil {
					return err
				}

				// If no rows updated, insert
				rows, _ := result.RowsAffected()
				if rows == 0 {
					_, err = tx.Exec(`
						INSERT INTO schema_metadata (schema_id, description, author, created_at, updated_at)
						VALUES ($1, $2, $3, NOW(), NOW())
					`, schema.ID, schema.Metadata.Description, schema.Metadata.Author)
					if err != nil {
						return err
					}
				}
			}
		case "deleted":
			_, err = tx.Exec(`
				DELETE FROM schema_metadata WHERE schema_id = $1
			`, schema.ID)
			if err != nil {
				return err
			}
		}
	}

	// Handle tags - delete all and recreate
	if diff.TagsChange != nil {
		_, err = tx.Exec(`DELETE FROM schema_tags WHERE schema_id = $1`, schema.ID)
		if err != nil {
			return err
		}

		for _, tag := range schema.Tags {
			_, err = tx.Exec(`
				INSERT INTO schema_tags (schema_id, tag_key, tag_value, created_at, updated_at)
				VALUES ($1, $2, $3, NOW(), NOW())
				`, schema.ID, tag.Key, tag.Value)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// DeleteSchema removes a schema (CASCADE handled by DB constraints)
func (s *DatabaseService) DeleteSchema(name string) error {
	result, err := s.db.Exec(`DELETE FROM schemas WHERE name = $1`, name)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("schema '%s' not found", name)
	}

	return nil
}
