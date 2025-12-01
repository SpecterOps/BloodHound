package load

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/specterops/bloodhound/packages/go/diffengine/internal/models"
)

type LoadService struct{}

type Load interface {
	LoadSchemasFromFile(filepath string) ([]*models.Schema, error)
	LoadSchemaByNameFromFile(filepath, schemaName string) (*models.Schema, error)
}

func NewLoadService() (Load, error) {
	return &LoadService{}, nil
}

// Top-level wrapper
type SchemaFiles struct {
	Schemas []SchemaFile `json:"schemas"`
}

// SchemaFile represents the JSON structure for schema files
type SchemaFile struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Fields   []FieldFile       `json:"fields"`
	Metadata *MetadataFile     `json:"metadata,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
}

type FieldFile struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsRequired bool   `json:"is_required"`
}

type MetadataFile struct {
	Description string `json:"description"`
	Author      string `json:"author"`
}

func (s *LoadService) LoadSchemasFromFile(filepath string) ([]*models.Schema, error) {
	var (
		schemaFiles SchemaFiles
	)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &schemaFiles); err != nil {
		return nil, err
	}

	schemas := make([]*models.Schema, 0, len(schemaFiles.Schemas))
	for _, schemaFile := range schemaFiles.Schemas {
		schema := convertSchemaFile(schemaFile)
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// LoadSchemaByNameFromFile loads a specific schema by name from a file
func (s *LoadService) LoadSchemaByNameFromFile(filepath, schemaName string) (*models.Schema, error) {
	schemas, err := s.LoadSchemasFromFile(filepath)
	if err != nil {
		return nil, err
	}

	for _, schema := range schemas {
		if schema.Name == schemaName {
			return schema, nil
		}
	}

	return nil, fmt.Errorf("schema '%s' not found in file", schemaName)
}

func convertSchemaFile(schemaFile SchemaFile) *models.Schema {
	// Convert to models.Schema
	schema := &models.Schema{
		Name:    schemaFile.Name,
		Version: schemaFile.Version,
		Fields:  make([]models.Field, 0, len(schemaFile.Fields)),
	}

	// Convert fields
	for _, f := range schemaFile.Fields {
		schema.Fields = append(schema.Fields, models.Field{
			Name:       f.Name,
			Type:       f.Type,
			IsRequired: f.IsRequired,
		})
	}

	// Convert metadata
	if schemaFile.Metadata != nil {
		schema.Metadata = &models.Metadata{
			Description: schemaFile.Metadata.Description,
			Author:      schemaFile.Metadata.Author,
		}
	}

	// Convert tags
	for key, value := range schemaFile.Tags {
		schema.Tags = append(schema.Tags, models.Tag{
			Key:   key,
			Value: value,
		})
	}

	return schema
}
