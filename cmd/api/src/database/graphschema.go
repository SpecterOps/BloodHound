package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"gorm.io/gorm"
)

type OpenGraphSchema interface {
	CreateGraphSchemaExtension(ctx context.Context, name string, displayName string, version string) (model.GraphSchemaExtension, error)
	GetGraphSchemaExtensionById(ctx context.Context, extensionId int) (model.GraphSchemaExtension, error)
}

func (s *BloodhoundDB) CreateGraphSchemaExtension(ctx context.Context, name string, displayName string, version string) (model.GraphSchemaExtension, error) {
	var (
		extension = model.GraphSchemaExtension{
			Name:        name,
			DisplayName: displayName,
			Version:     version,
		}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateGraphSchemaExtension,
			Model:  &extension, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		if result := tx.Raw(fmt.Sprintf(`
			INSERT INTO %s (name, display_name, version, created_at)
			VALUES (?, ?, ?, NOW())
			RETURNING id, name, display_name, version, created_at`,
			extension.TableName()),
			name, displayName, version).Scan(&extension); result.Error != nil {
			if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
				return fmt.Errorf("%w: %v", ErrDuplicateGraphSchemaExtensionName, result.Error)
			}
			return CheckError(result)
		}
		return nil
	}); err != nil {
		return model.GraphSchemaExtension{}, err
	}

	return extension, nil
}

func (s *BloodhoundDB) GetGraphSchemaExtensionById(ctx context.Context, extensionId int) (model.GraphSchemaExtension, error) {
	var (
		extension = model.GraphSchemaExtension{
			ID: extensionId,
		}
	)

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if result := tx.Raw(fmt.Sprintf(`
			SELECT id, name, display_name, version, is_builtin, created_at, updated_at, deleted_at
			FROM %s WHERE id = ?`,
			extension.TableName()),
			extensionId).First(&extension); result.Error != nil {
			return CheckError(result)
		}
		return nil
	}); err != nil {
		return model.GraphSchemaExtension{}, err
	}

	return extension, nil
}
