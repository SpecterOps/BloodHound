package graph

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"os"
	"path/filepath"
	"time"

	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"github.com/specterops/bloodhound/src/migrations"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/graphify"
	"github.com/specterops/bloodhound/src/services/upload"
)

const (
	Name  = "graph"
	Usage = "Run code generation in current workspace"
)

type command struct {
	env environment.Environment
}

// Create new instance of command to capture given environment
func Create(env environment.Environment) *command {
	return &command{
		env: env,
	}
}

// Usage of command
func (s *command) Usage() string {
	return Usage
}

// Name of command
func (s *command) Name() string {
	return Name
}

// Parse command flags
func (s *command) Parse(cmdIndex int) error {
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	return nil
}

// Run generate command
func (s *command) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// connect to database
	database, err := initializeDatabase(ctx)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	// loads the generic ingest schema from node/edge json files
	schema, err := upload.LoadIngestSchema()
	if err != nil {
		return fmt.Errorf("error loading schema %v", err)
	}

	// get test data path
	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("error finding workspace root: %w", err)
	}

	ingestFilePath := filepath.Join(paths.Root + "/cmd/api/src/test/fixtures/fixtures/v6/ingest/")

	// return all files in test data path directory
	ingestFiles, err := os.ReadDir(ingestFilePath)
	if err != nil {
		return fmt.Errorf("error reading ingest directory: %v", err)
	}

	var errs []error

	for _, entry := range ingestFiles {
		err := database.BatchOperation(ctx, func(batch graph.Batch) error {
			timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())

			readOpts := graphify.ReadOptions{IngestSchema: schema, FileType: model.FileTypeJson, ADCSEnabled: true}

			file, err := os.Open(ingestFilePath + "/" + entry.Name())
			if err != nil {
				return fmt.Errorf("error opening JSON file %s: %v", entry.Name(), err)
			}
			defer file.Close()

			// ingest file into database
			if err := graphify.ReadFileForIngest(timestampedBatch, file, readOpts); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					errs = append(errs, fmt.Errorf("error ingesting timestamped batch %v: /n error: %w", timestampedBatch, sql.ErrNoRows))
				}
				errs = append(errs, fmt.Errorf("error ingesting timestamped batch %v: error: %w", timestampedBatch, err))
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("error occurred during batch operation %w", err)
		}
	}

	// TODO: log ingest errors

	// Read nodes and edges from database
	nodes, edges, err := getNodesAndEdges(database)
	if err != nil {
		return fmt.Errorf("error retrieving nodes and edges from database %w", err)
	}

	// transform to arrows.app files
	err = transformToArrows(nodes, edges)
	if err != nil {
		return fmt.Errorf("error transforming nodes and edges to arrows.app format %w", err)
	}

	return nil
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Everything in arrows.app is strings
type Node struct {
	ID       string   `json:"id"`
	Position Position `json:"position"`
	// "caption": "", // name of the node — if we have a name, we use the name, if not, use object id
	// "labels": [], // kinds that we get back from database — loop through and convert to a string
	// "properties": {}, map[string]string - map[string]any = typecast to strings

}

type Relationship struct {
	ID   string `json:"id"`
	// type: label - relationship kinds - not an id but kind is a value
	From string `json:"fromId"`
	To   string `json:"toId"`
	// same deal for properties
}

type Arrows struct {
	Nodes         []Node         `json:"nodes"`
	Relationships []Relationship `json:"relationships"`
}

func transformToArrows(nodes []*graph.Node, edges []*graph.Relationship) error {
	var arrowNodes []Node
	var arrowEdges []Relationship

	for i := range nodes {
		arrowNodes = append(arrowNodes, Node{
			ID: fmt.Sprint(i),
			Position: Position{
				X: 0,
				Y: 0,
			},
		})
	}

	for _, e := range edges {
		arrowEdges = append(arrowEdges, Relationship{
			ID:   e.ID.String(),
			From: e.StartID.String(),
			To:   e.EndID.String(),
		})
	}

	arrows := Arrows{
		Nodes:         arrowNodes,
		Relationships: arrowEdges,
	}

	jsonBytes, err := json.MarshalIndent(arrows, "", "  ")
	if err != nil {
		return err
	}

	// TODO: make this agnostic - allow for user to define where this should live
	return os.WriteFile("arrows.json", jsonBytes, 0644)
}

func getNodesAndEdges(database graph.Database) ([]*graph.Node, []*graph.Relationship, error) {
	var nodes []*graph.Node
	var edges []*graph.Relationship
	err := database.ReadTransaction(context.TODO(), func(tx graph.Transaction) error {
		err := tx.Nodes().Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for node := range cursor.Chan() {
				nodes = append(nodes, node)
			}
			return cursor.Error()
		})
		if err != nil {
			return fmt.Errorf("error fetching nodes %w", err)
		}
		err = tx.Relationships().Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for edge := range cursor.Chan() {
				edges = append(edges, edge)
			}
			return cursor.Error()
		})
		if err != nil {
			return fmt.Errorf("error fetching relationships %w", err)
		}

		return nil
	})

	if err != nil {
		return nodes, edges, fmt.Errorf("error occurred reading the database %w", err)
	}
	return nodes, edges, nil
}

func initializeDatabase(ctx context.Context) (graph.Database, error) {
	// TODO: use an sb_environment variable - or a flag on subcommand
	connection := "user=bloodhound password=bloodhoundcommunityedition dbname=bloodhound host=localhost port=65432"
	if pool, err := pg.NewPool(connection); err != nil {
		return nil, err
	} else {
		database, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
			GraphQueryMemoryLimit: size.Gibibyte,
			ConnectionString:      connection,
			Pool:                  pool,
		})
		if err != nil {
			return nil, err
		}

		migrator := migrations.NewGraphMigrator(database)
		err = migrator.Migrate(ctx, graphschema.DefaultGraphSchema())
		if err != nil {
			return nil, err
		}

		err = database.SetDefaultGraph(ctx, graphschema.DefaultGraph())
		if err != nil {
			return nil, err
		}

		return database, nil
	}
}
