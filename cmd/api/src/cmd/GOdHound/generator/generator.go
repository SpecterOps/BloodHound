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
	//Kind    string `json:"kind"`
	Value string `json:"value"`
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
