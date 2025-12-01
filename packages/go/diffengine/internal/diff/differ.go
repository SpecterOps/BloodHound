package diff

import (
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/diffengine/internal/models"
)

type DiffService struct{}

type Diff interface {
	DiffSchemas(existing *models.Schema, incoming *models.Schema) *DiffResult
	PrintAnalysis(diffResult *DiffResult)
}

func NewDiffService() (Diff, error) {
	return &DiffService{}, nil
}

type DiffResult struct {
	IsNew          bool
	FieldsToAdd    []models.Field
	FieldsToUpdate []FieldUpdate
	FieldsToDelete []models.Field
	MetadataChange *MetadataChange
	TagsChange     *TagsChange
}

// FieldUpdate captures what changed in a field
type FieldUpdate struct {
	Name        string
	OldType     string
	NewType     string
	OldRequired bool
	NewRequired bool
}

// MetadataChange captures metadata changes
type MetadataChange struct {
	Action         string // "created", "updated", "deleted"
	OldDescription string
	NewDescription string
	OldAuthor      string
	NewAuthor      string
}

// TagsChange captures tag changes
type TagsChange struct {
	Added   []TagChange
	Removed []TagChange
	Updated []TagChange
}

type TagChange struct {
	Key      string
	OldValue string
	NewValue string
}

// DiffSchemas compares incoming schema with existing DB schema
func (s *DiffService) DiffSchemas(existing *models.Schema, incoming *models.Schema) *DiffResult {
	if existing == nil {
		return &DiffResult{IsNew: true}
	}

	result := &DiffResult{}

	// Build maps for comparison using natural keys (field names)
	existingFields := make(map[string]models.Field)
	for _, f := range existing.Fields {
		existingFields[f.Name] = f
	}

	incomingFields := make(map[string]models.Field)
	for _, f := range incoming.Fields {
		incomingFields[f.Name] = f
	}

	// Find fields to add and update
	for name, incomingField := range incomingFields {
		if existingField, exists := existingFields[name]; exists {
			// Field exists - check if it changed
			if existingField.Type != incomingField.Type ||
				existingField.IsRequired != incomingField.IsRequired {
				result.FieldsToUpdate = append(result.FieldsToUpdate, FieldUpdate{
					Name:        name,
					OldType:     existingField.Type,
					NewType:     incomingField.Type,
					OldRequired: existingField.IsRequired,
					NewRequired: incomingField.IsRequired,
				})
			}
		} else {
			// Field doesn't exist - add it
			result.FieldsToAdd = append(result.FieldsToAdd, incomingField)
		}
	}

	// Find fields to delete
	for name, existingField := range existingFields {
		if _, exists := incomingFields[name]; !exists {
			result.FieldsToDelete = append(result.FieldsToDelete, existingField)
		}
	}

	// Check metadata changes
	result.MetadataChange = analyzeMetadataChange(existing.Metadata, incoming.Metadata)

	// Check tags changes
	result.TagsChange = analyzeTagsChange(existing.Tags, incoming.Tags)

	return result
}

func analyzeMetadataChange(existing *models.Metadata, incoming *models.Metadata) *MetadataChange {
	// No change
	if existing == nil && incoming == nil {
		return nil
	}

	// Metadata created
	if existing == nil && incoming != nil {
		return &MetadataChange{
			Action:         "created",
			NewDescription: incoming.Description,
			NewAuthor:      incoming.Author,
		}
	}

	// Metadata deleted
	if existing != nil && incoming == nil {
		return &MetadataChange{
			Action:         "deleted",
			OldDescription: existing.Description,
			OldAuthor:      existing.Author,
		}
	}

	// Metadata updated
	if existing.Description != incoming.Description || existing.Author != incoming.Author {
		return &MetadataChange{
			Action:         "updated",
			OldDescription: existing.Description,
			NewDescription: incoming.Description,
			OldAuthor:      existing.Author,
			NewAuthor:      incoming.Author,
		}
	}

	return nil
}

func analyzeTagsChange(existing, incoming []models.Tag) *TagsChange {
	existingMap := make(map[string]string) // key -> value
	incomingMap := make(map[string]string) // key -> value

	for _, tag := range existing {
		existingMap[tag.Key] = tag.Value
	}

	for _, tag := range incoming {
		incomingMap[tag.Key] = tag.Value
	}

	change := &TagsChange{
		Added:   []TagChange{},
		Removed: []TagChange{},
		Updated: []TagChange{},
	}

	// Find added and updated tags
	for key, newValue := range incomingMap {
		if oldValue, exists := existingMap[key]; exists {
			// Tag key exists - check if value changed
			if oldValue != newValue {
				change.Updated = append(change.Updated, TagChange{
					Key:      key,
					OldValue: oldValue,
					NewValue: newValue,
				})
			}
		} else {
			// Tag key doesn't exist - it's new
			change.Added = append(change.Added, TagChange{
				Key:      key,
				NewValue: newValue,
			})
		}
	}

	// Find removed tags
	for key, oldValue := range existingMap {
		if _, exists := incomingMap[key]; !exists {
			change.Removed = append(change.Removed, TagChange{
				Key:      key,
				OldValue: oldValue,
			})
		}
	}

	// If no changes, return nil
	if len(change.Added) == 0 && len(change.Removed) == 0 && len(change.Updated) == 0 {
		return nil
	}

	return change
}

func (s *DiffService) PrintAnalysis(diffResult *DiffResult) {
	logger := slog.Default()

	logger.Info("=======================================================")
	logger.Info("***************** DIFF ANALYSIS ***********************")
	logger.Info("=======================================================")

	fieldsLog := logger.WithGroup("fields")

	fieldsLog.Info("FIELDS ANALYSIS")
	fieldsLog.Info("-------------------------------------------------------")

	if len(diffResult.FieldsToAdd) > 0 {
		addGroup := fieldsLog.WithGroup("added")
		addGroup.Info("Fields added",
			slog.Int("count", len(diffResult.FieldsToAdd)),
		)

		for _, f := range diffResult.FieldsToAdd {
			addGroup.Info("Added field",
				slog.String("name", f.Name),
				slog.String("type", f.Type),
				slog.Bool("required", f.IsRequired),
			)
		}
	}

	if len(diffResult.FieldsToUpdate) > 0 {
		updateGroup := fieldsLog.WithGroup("modified")
		updateGroup.Info("Fields modified",
			slog.Int("count", len(diffResult.FieldsToUpdate)),
		)

		for _, f := range diffResult.FieldsToUpdate {
			fieldGroup := updateGroup.With(slog.String("field", f.Name))

			if f.OldType != f.NewType {
				fieldGroup.Info("Type changed",
					slog.String("old", f.OldType),
					slog.String("new", f.NewType),
				)
			}

			if f.OldRequired != f.NewRequired {
				fieldGroup.Info("Required changed",
					slog.Bool("old", f.OldRequired),
					slog.Bool("new", f.NewRequired),
				)
			}
		}
	}

	if len(diffResult.FieldsToDelete) > 0 {
		delGroup := fieldsLog.WithGroup("deleted")
		delGroup.Info("Fields deleted",
			slog.Int("count", len(diffResult.FieldsToDelete)),
		)

		for _, f := range diffResult.FieldsToDelete {
			delGroup.Info("Deleted field",
				slog.String("name", f.Name),
				slog.String("type", f.Type),
			)
		}
	}

	if len(diffResult.FieldsToAdd) == 0 &&
		len(diffResult.FieldsToUpdate) == 0 &&
		len(diffResult.FieldsToDelete) == 0 {
		fieldsLog.Info("No field changes")
	}

	metaLog := logger.WithGroup("metadata")
	metaLog.Info("-------------------------------------------------------")
	metaLog.Info("METADATA ANALYSIS")
	metaLog.Info("-------------------------------------------------------")

	if diffResult.MetadataChange != nil {
		m := diffResult.MetadataChange

		switch m.Action {
		case "created":
			metaLog.Info("Metadata created",
				slog.String("description", m.NewDescription),
				slog.String("author", m.NewAuthor),
			)

		case "updated":
			metaLog.Info("Metadata updated")

			if m.OldDescription != m.NewDescription {
				metaLog.Info("Description changed",
					slog.String("old", m.OldDescription),
					slog.String("new", m.NewDescription),
				)
			}
			if m.OldAuthor != m.NewAuthor {
				metaLog.Info("Author changed",
					slog.String("old", m.OldAuthor),
					slog.String("new", m.NewAuthor),
				)
			}

		case "deleted":
			metaLog.Info("Metadata deleted",
				slog.String("description", m.OldDescription),
				slog.String("author", m.OldAuthor),
			)
		}
	} else {
		metaLog.Info("No metadata changes")
	}

	tagsLog := logger.WithGroup("tags")
	tagsLog.Info("-------------------------------------------------------")
	tagsLog.Info("TAGS ANALYSIS")
	tagsLog.Info("-------------------------------------------------------")

	if diffResult.TagsChange != nil {
		tc := diffResult.TagsChange

		if len(tc.Added) > 0 {
			add := tagsLog.WithGroup("added")
			add.Info("Tags added", slog.Int("count", len(tc.Added)))
			for _, tag := range tc.Added {
				add.Info("Added tag",
					slog.String("key", tag.Key),
					slog.String("value", tag.NewValue),
				)
			}
		}

		if len(tc.Updated) > 0 {
			upd := tagsLog.WithGroup("modified")
			upd.Info("Tags modified", slog.Int("count", len(tc.Updated)))
			for _, tag := range tc.Updated {
				upd.Info("Modified tag",
					slog.String("key", tag.Key),
					slog.String("old", tag.OldValue),
					slog.String("new", tag.NewValue),
				)
			}
		}

		if len(tc.Removed) > 0 {
			rem := tagsLog.WithGroup("removed")
			rem.Info("Tags removed", slog.Int("count", len(tc.Removed)))
			for _, tag := range tc.Removed {
				rem.Info("Removed tag",
					slog.String("key", tag.Key),
					slog.String("value", tag.OldValue),
				)
			}
		}

	} else {
		tagsLog.Info("No tag changes")
	}

	logger.Info("-------------------------------------------------------")
}
