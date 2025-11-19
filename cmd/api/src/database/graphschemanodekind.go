package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

const (
	SchemaNodeKindsTableName = "schema_node_kinds"
)

// CreateSchemaNodeKind - Creates a new kind record for the provided extension. The populated model.SchemaNodeKind is returned.
func (s *BloodhoundDB) CreateSchemaNodeKind(ctx context.Context, name string, extensionID int32, displayName string, description string, isDisplayKind bool, icon, iconColor string) (model.SchemaNodeKind, error) {
	graphSchemaNodeKind := model.SchemaNodeKind{
		Name:              name,
		SchemaExtensionId: extensionID,
		DisplayName:       displayName,
		Description:       description,
		IsDisplayKind:     isDisplayKind,
		Icon:              icon,
		IconColor:         iconColor,
	}

	if result := s.db.WithContext(ctx).Create(&graphSchemaNodeKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return model.SchemaNodeKind{}, fmt.Errorf("%w: %v", ErrDuplicateSchemaNodeKindName, result.Error)
		}
		return model.SchemaNodeKind{}, CheckError(result)
	}
	return graphSchemaNodeKind, nil
}

// GetSchemaNodeKindByID - retrieves the model.SchemaNodeKind based on the provided id. If no record exists for the id, it returns an error
func (s *BloodhoundDB) GetSchemaNodeKindByID(ctx context.Context, schemaNodeKindID int32) (model.SchemaNodeKind, error) {
	var graphSchemaNodeKind model.SchemaNodeKind
	return graphSchemaNodeKind, CheckError(s.db.WithContext(ctx).Table(SchemaNodeKindsTableName).Where("id = ?", schemaNodeKindID).First(&graphSchemaNodeKind))
}
