package bloodhound

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/services/ingest"
)

type App struct {
	db      database.Database
	graphDB graph.Database
	cfg     config.Configuration

	// Any number of exported services that are containers to pin methods on
	// without polluting app into another nasty mess
	IngestService ingest.IngestService
}

func NewApp(db database.Database, graphDB graph.Database, cfg config.Configuration) App {
	app := App{
		db:      db,
		graphDB: graphDB,
		cfg:     cfg,

		// Initialize each service with the necessary dependancies
		// TODO: Determine if the ingest service needs all of these?
		IngestService: ingest.NewIngestService(db, graphDB, cfg),
	}

	return app
}
