package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	db "github.com/specterops/bloodhound/packages/go/diffengine/internal/database"
	"github.com/specterops/bloodhound/packages/go/diffengine/internal/models"
	"gorm.io/gorm"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Error running diff engine", slog.Any("err", err))
        os.Exit(1)
    }
}

func run() error {
	var (
		connection         = os.Getenv("CONNECTION_STRING")
		originalSchemaPath = os.Getenv("OG_SCHEMA_PATH")
		newSchemaPath      = os.Getenv("NEW_SCHEMA_PATH")
		schemaName         = os.Getenv("SCHEMA_NAME")
	)

	if db, err := initializeDatabase(connection); err != nil {
		return fmt.Errorf("error setting up database: %v", err)
	} else if services, err := servicesSetup(db); err != nil {
		return fmt.Errorf("error setting up services: %v", err)
	} else if v1Schema, err := services.LoadService.LoadSchemaByNameFromFile(originalSchemaPath, schemaName); err != nil {
		return fmt.Errorf("error loading original schema name %s from file: %v", schemaName, err)
	} else if existingSchema, err := services.DBService.FetchSchemaByName(v1Schema.Name); err != nil {
		return fmt.Errorf("error fetching schema by name %s: %v", schemaName, err)
	} else if err := createSchema(existingSchema, v1Schema, services.DBService); err != nil {
		return fmt.Errorf("error checking schema: %v", err)
	} else if v2Schema, err := services.LoadService.LoadSchemaByNameFromFile(newSchemaPath, schemaName); err != nil {
		return fmt.Errorf("error loading new schema name %s from file: %v", schemaName, err)
	} else {
		services.DiffService.PrintAnalysis(services.DiffService.DiffSchemas(existingSchema, v2Schema))
		return nil
	}
}

func initializeDatabase(connection string) (*gorm.DB, error) {
	gormDB, err := database.OpenDatabase(connection)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	return gormDB, nil
}

func createSchema(schemaFromDatabase *models.Schema, schema *models.Schema, database db.DatabaseOperations) error {
	if schemaFromDatabase == nil {
		slog.Info("Schema doesn't exist. Creating new schema...")
		if err := database.CreateSchema(schema); err != nil {
			return err
		}

		slog.Info("Schema created successfully")
	} else {
		slog.Info("Schema already exists. Skipping creation.")
	}

	return nil
}
