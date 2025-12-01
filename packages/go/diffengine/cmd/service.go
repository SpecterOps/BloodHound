package main

import (
	"fmt"

	db "github.com/specterops/bloodhound/packages/go/diffengine/internal/database"
	"github.com/specterops/bloodhound/packages/go/diffengine/internal/diff"
	"github.com/specterops/bloodhound/packages/go/diffengine/internal/load"
	"gorm.io/gorm"
)

type Services struct {
	DBService   db.DatabaseOperations
	DiffService diff.Diff
	LoadService load.Load
}

func servicesSetup(gormDB *gorm.DB) (*Services, error) {
	if dbService, err := db.NewDatabaseService(gormDB); err != nil {
		return nil, fmt.Errorf("error setting up diff service: %w", err)
	} else if diffService, err := diff.NewDiffService(); err != nil {
		return nil, fmt.Errorf("error setting up diff service: %w", err)
	} else if loadService, err := load.NewLoadService(); err != nil {
		return nil, fmt.Errorf("error setting up diff service: %w", err)
	} else {
		return &Services{
			DBService:   dbService,
			DiffService: diffService,
			LoadService: loadService,
		}, nil
	}
}
