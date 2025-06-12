package generator

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
)

type GenericIngestNode struct {
	ID         string         `json:"id"`
	Kinds      []string       `json:"kinds"`
	Properties map[string]any `json:"properties"`
}

type GenericIngestEndpoint struct {
	MatchBy string `json:"match_by"`
	Kind    string `json:"kind"`
	Value   string `json:"value"`
}

type GenericIngestEdge struct {
	Kind       string                `json:"kind"`
	Start      GenericIngestEndpoint `json:"start"`
	End        GenericIngestEndpoint `json:"end"`
	Properties map[string]any        `json:"properties"`
}

type GenericIngestGraphContent struct {
	Nodes []GenericIngestNode `json:"nodes"`
	Edges []GenericIngestEdge `json:"edges"`
}

type GenericIngestGraph struct { //Graph
	Graph GenericIngestGraphContent `json:"graph"`
}

type GraphNode struct {
	ID         string                 `json:"id"`
	Kinds      []string               `json:"kinds"`
	Properties map[string]interface{} `json:"properties"`
}

/*
	func MakeDomain() (GenericIngestPayload, error) {
		domain := GenericIngestNode{
			ID:    "1234",
			Kinds: []string{"Domain", "Base"},
			Properties: map[string]any{
				// This MUST match the ID value of the GenericIngestNode otherwise it will fail validation
				"objectid": "1234",
			},
		}

		return GenericIngestPayload{
			Graph: GenericIngestGraph{
				Nodes: []GenericIngestNode{domain},
			},
		}, nil
	}
*/
func getRandomName(r *rand.Rand, filePath string) string {
	file, err := os.Open(filePath)

	if err != nil {
		log.Fatalf("❌ Could not open %s: %v", filePath, err)
	}
	defer file.Close()

	var names []string
	if err := json.NewDecoder(file).Decode(&names); err != nil {
		log.Fatalf("❌ Could not decode JSON from %s: %v", filePath, err)
	}

	return names[r.Intn(len(names))]
}

func CreateHeader(filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("❌ Could not create file %s: %v", filePath, err)
	}
	defer file.Close()

	header := `{
  "graph": {
`

	if _, err := file.WriteString(header); err != nil {
		log.Fatalf("❌ Could not write header to %s: %v", filePath, err)
	}
}

func AddNodes(filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("❌ Could not open file %s: %v", filePath, err)
	}
	defer file.Close()

	node := ""
	//write section header
	line := `    "nodes": [`
	if _, err := file.WriteString(line); err != nil {
		log.Fatalf("❌ Could not write header to %s: %v", filePath, err)
	}
	// insert all Nodes
	node = `      {
        "id": "00000001",
        "kinds": [
          "User",
          "Base"
        ],
        "properties": {
          "displayname": "Scoubi Dou",
          "property": "a",
          "objectid": "00000001",
          "name": "SCOUBI"
        }
      },
      {
        "id": "00000002",
        "kinds": [
          "Group",
          "Base"
        ],
        "properties": {
          "displayname": "Scoubi's Computer",
          "property": "a",
          "objectid": "00000002",
          "name": "Padawan-GOdev"
        }
      }	  
		`
	if _, err := file.WriteString(node); err != nil {
		log.Fatalf("❌ Could not write header to %s: %v", filePath, err)
	}
	//write section footer
	line = `    ],
	`
	if _, err := file.WriteString(line); err != nil {
		log.Fatalf("❌ Could not write header to %s: %v", filePath, err)
	}
}

/*
func AddEdges(filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("❌ Could not open file %s: %v", filePath, err)
	}
	defer file.Close()

	edge := ""
	//write section header
	line := `    "edges": [
	`
	if _, err := file.WriteString(line); err != nil {
		log.Fatalf("❌ Could not write footer to %s: %v", filePath, err)
	}
	// insert all Edges
	edge = `      {
        "kind": "MemberOf",
        "start": {
          "value": "00000001",
          "match_by": "id"
        },
        "end": {
          "value": "00000002",
          "match_by": "id"
        },
        "properties": {}
      }
	`
	if _, err := file.WriteString(edge); err != nil {
		log.Fatalf("❌ Could not write footer to %s: %v", filePath, err)
	}
	//write section footer
	line = `    ]
  },
  `
	if _, err := file.WriteString(line); err != nil {
		log.Fatalf("❌ Could not write footer to %s: %v", filePath, err)
	}
}

func CreateFooter(filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("❌ Could not create file %s: %v", filePath, err)
	}
	defer file.Close()

	// Check file size
	info, err := file.Stat()
	if err != nil {
		log.Fatalf("❌ Could not stat file %s: %v", filePath, err)
	}

	if info.Size() == 0 {
		log.Printf("⚠️  File %s is empty — skipping footer write.\n", filePath)
		return
	}

	footer := `
	  "metadata": {
    "ingest_version": "v1",
    "collector": {
      "name": "GOdHound",
      "version": "beta",
      "properties": {
        "collection_methods": [
          "GOdHound"
        ],
        "windows_server_version": "n/a"
      }
    }
  }
}
`

	if _, err := file.WriteString(footer); err != nil {
		log.Fatalf("❌ Could not write footer to %s: %v", filePath, err)
	}
}
*/
