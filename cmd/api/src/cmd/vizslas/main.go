package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/cmd/vizslas/golang"
	"github.com/specterops/bloodhound/src/cmd/vizslas/ingest"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/version"
	"os"
)

const (
	workspaceDir = "/home/zinic/work/bhe_copy"
)

func printVersion() {
	fmt.Printf("Vizslas %s\n", version.GetVersion())
	os.Exit(0)
}

func main() {
	log.ConfigureDefaults()

	var (
		configFilePath string
		logFilePath    string
		versionFlag    bool
	)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Vizslas for Golang\n\nUsage of %s\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&versionFlag, "version", false, "Get binary version.")
	flag.StringVar(&configFilePath, "configfile", bootstrap.DefaultConfigFilePath(), "Configuration file to load.")
	flag.StringVar(&logFilePath, "logfile", config.DefaultLogFilePath, "Log file to write to.")
	flag.Parse()

	if versionFlag {
		printVersion()
	}

	// Initialize basic logging facilities while we start up
	log.ConfigureDefaults()

	ctx := bootstrap.NewDaemonContext(context.Background())

	if cfg, err := config.GetConfiguration(configFilePath, config.NewDefaultConfiguration); err != nil {
		log.Fatalf("Unable to read configuration %s: %v", configFilePath, err)
	} else if graphDB, err := dawgs.Open(ctx, neo4j.DriverName, dawgs.Config{
		DriverCfg: cfg.Neo4J.Neo4jConnectionString(),
	}); err != nil {
		log.Fatalf("Unable to connect to graph %s: %v", configFilePath, err)
	} else {
		defer graphDB.Close(ctx)

		if err := os.Setenv("CGO_ENABLED", "0"); err != nil {
			log.Fatalf("Failed to set environment variable: %v", err)
		}

		if ingestPayload, err := golang.AnalyzeGoWorkspace(ctx, workspaceDir, "/home/zinic/main.zip", "/home/zinic/vulndb-latest.zip"); err != nil {
			log.Errorf("Failed to list workspace at %s: %v", workspaceDir, err)
		} else {
			if err := graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
				for _, visitedNode := range ingestPayload.Visited.Nodes {
					if err := ingest.CheckEntity(visitedNode); err != nil {
						return err
					}

					var (
						idKind = graph.StringKind(visitedNode.Kind)

						update = graph.NodeUpdate{
							Node: graph.PrepareNode(
								graph.AsProperties(visitedNode.Properties),
								append(graph.Kinds{idKind}, graph.StringsToKinds(visitedNode.ExtendedKinds)...)...,
							),
							IdentityKind:       idKind,
							IdentityProperties: visitedNode.IDKeys,
						}
					)

					if err := batch.UpdateNodeBy(update); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				log.Fatalf("Failed to ingest Golang workspace: %v", err)
			}

			if err := graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
				for _, visitedEdge := range ingestPayload.Visited.Edges {
					if err := ingest.CheckEntity(visitedEdge.Start); err != nil {
						return err
					}

					if err := ingest.CheckEntity(visitedEdge.End); err != nil {
						return err
					}

					update := graph.RelationshipUpdate{
						Relationship: &graph.Relationship{
							Kind:       graph.StringKind(visitedEdge.Kind),
							Properties: graph.AsProperties(visitedEdge.Properties),
						},

						Start: graph.PrepareNode(
							graph.AsProperties(visitedEdge.Start.Properties),
							graph.StringKind(visitedEdge.Start.Kind),
						),
						StartIdentityKind:       graph.StringKind(visitedEdge.Start.Kind),
						StartIdentityProperties: visitedEdge.Start.IDKeys,

						End: graph.PrepareNode(
							graph.AsProperties(visitedEdge.End.Properties),
							graph.StringKind(visitedEdge.End.Kind),
						),
						EndIdentityKind:       graph.StringKind(visitedEdge.End.Kind),
						EndIdentityProperties: visitedEdge.End.IDKeys,
					}

					if err := batch.UpdateRelationshipBy(update); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				log.Fatalf("Failed to ingest Golang workspace: %v", err)
			}
		}

		//if written, err := golang.DownloadGoVulnDatabaseArchive(workingDir); err != nil {
		//	log.Errorf("Failed to download Golang Vulnerability database: %v", err)
		//} else {
		//	log.Infof("Wrote %d bytes to %s", written, workingDir)
		//}
	}
}
