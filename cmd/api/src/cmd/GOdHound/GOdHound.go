// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/specterops/bloodhound/src/cmd/GOdHound/client"
	"github.com/specterops/bloodhound/src/cmd/GOdHound/config"
	"github.com/specterops/bloodhound/src/cmd/GOdHound/generator"
)

func fatal(value any) {
	fmt.Printf("%v", value)
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fatal(fmt.Sprintf(format, args...))
}

func uploadFiles(dir string) {
	var (
		token = &client.TokenCredentialsHandler{
			TokenID:  config.Server.APIID,
			TokenKey: config.Server.APIKey,
		}
	)

	if apiClient, err := client.NewClient(config.Server.URL, token); err != nil {
		fatalf("Failed to create client: %v", err)
	} else {
		if version, err := apiClient.Version(); err != nil {
			fatalf("Failed to get version: %v", err)
		} else {
			fmt.Printf("Version: %s\n", version)
		}

		if err != nil {
			fatal(err)
		}

		if newFileUploadJob, err := apiClient.StartFileUploadJob(); err != nil {
			fatal(err)
		} else {
			//dir := ("./output/ad_sample/")

			entries, err := os.ReadDir(dir)
			if err != nil {
				log.Fatalf("❌ Failed to read directory: %v", err)
			}
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
					continue
				}

				fullPath := filepath.Join(dir, entry.Name())
				// Open and process the file
				file, err := os.Open(fullPath)
				if err != nil {
					log.Printf("⚠️  Skipping file %s due to error: %v", fullPath, err)
					continue
				}

				if fin, err := os.Open(fullPath); err != nil {
					fatal(err)
				} else {
					if err := apiClient.SendFileUploadPart(newFileUploadJob, fin); err != nil {
						fatal(err)
					}
					fmt.Printf("File: %v uploaded", fin.Name())
					fin.Close()
				}

				if err := apiClient.EndFileUploadJob(newFileUploadJob); err != nil {
					fatal(err)
				}
				file.Close()
			} // FOR
		}
	}
}

func generateADGraph(r *rand.Rand) ([]generator.GenericIngestNode, []generator.GenericIngestEdge) {
	config.ConfDomain()

	nbUsers := 100
	nbComps := 100
	nbGroups := nbUsers / 10
	nbGPOs := nbUsers / 10
	nbOUs := nbUsers / 15
	nbEdges := nbUsers * nbComps * nbGroups * nbGPOs * nbOUs / 20000

	// Arrays that contain respectively all the Nodes & Edges we want to create
	var generatedNodes []generator.GenericIngestNode
	var generatedEdges []generator.GenericIngestEdge

	// Generate the Domain Object
	newObj := generator.AddADDomain()
	generatedNodes = append(generatedNodes, newObj.ToGraphNode())

	// Generate the Users
	for i := 1; i <= nbUsers; i++ {
		newObj := generator.AddADUser(r, int32(i), "")
		generatedNodes = append(generatedNodes, newObj.ToGraphNode())
	}

	// Generate the Service Account
	for i := 1; i <= nbUsers/3; i++ {
		newObj := generator.AddADUser(r, int32(len(generatedNodes)), "SVC")
		generatedNodes = append(generatedNodes, newObj.ToGraphNode())
	}

	firstComp := len(generatedNodes)
	// Generate the Endpoints
	for i := 1; i <= nbComps; i++ {
		newObj := generator.AddADComputer(r, int32(len(generatedNodes)), "WORKSTATION")
		generatedNodes = append(generatedNodes, newObj.ToGraphNode())
	}

	// Generate Servers
	for i := 1; i <= nbComps/2; i++ {
		newObj := generator.AddADComputer(r, int32(len(generatedNodes)), "SERVER")
		generatedNodes = append(generatedNodes, newObj.ToGraphNode())
	}
	lastComp := len(generatedNodes)

	//Generate Groups and add members to them
	for i := 1; i <= nbGroups; i++ {
		newObj := generator.AddADGoup(int32(len(generatedNodes)), "", "") // We don't send neither a name or a rid because we want a random one.
		generatedNodes = append(generatedNodes, newObj.ToGraphNode())

		// Add Members to the group we just created
		eNode := generatedNodes[len(generatedNodes)-1].ID              // End Node is that last group we've added.
		nbMembers := int(float64(nbUsers) * (0.10 + r.Float64()*0.05)) // 10% to 15%
		forCPU := r.Intn(3)                                            // We want ~25% of our group to be for computers and not users
		if forCPU == 0 {
			for y := 0; y < nbMembers; y++ {
				x := firstComp + int(r.Intn(int(lastComp-firstComp+1))) // Select a random computer object
				sNode := generatedNodes[x].ID
				newEdge := generator.AddADEdge(r, "MemberOf", sNode, eNode) // r & 0 won't be used, they are place holder
				generatedEdges = append(generatedEdges, newEdge.ToGraphEdge())
			}
		} else {
			for y := 0; y < nbMembers; y++ {
				x := r.Intn(nbUsers) + 1
				sNode := generatedNodes[x].ID
				newEdge := generator.AddADEdge(r, "MemberOf", sNode, eNode) // r & 0 won't be used, they are place holder
				generatedEdges = append(generatedEdges, newEdge.ToGraphEdge())
			}
		}
	}

	for i := 1; i <= nbGPOs; i++ {
		newObj := generator.AddADGPO(r, int32(len(generatedNodes)))
		generatedNodes = append(generatedNodes, newObj.ToGraphNode())
	}

	for i := 1; i <= nbOUs; i++ {
		newObj := generator.AddADOU(r, int32(len(generatedNodes)))
		generatedNodes = append(generatedNodes, newObj.ToGraphNode())
	}

	// Add Mandatory Groups

	newNode := generator.AddADGoup(int32(len(generatedNodes)), "DOMAIN ADMINS", "512") // We don't send a name because we want a random one.
	generatedNodes = append(generatedNodes, newNode.ToGraphNode())
	sNode := generatedNodes[len(generatedNodes)-1].ID // -1 because the array starts at 0 and we want to access the last Node we created.
	eNode := generatedNodes[0].ID                     // This should be our Domain node as it's the first object we create
	rel := []string{"AllExtendedRights", "GenericWrite", "WriteDACL", "WriteOwner", "WriteOwnerRaw"}

	for _, eType := range rel {
		newEdge := generator.AddADEdge(r, eType, sNode, eNode) // r & 0 won't be used, they are place holder
		generatedEdges = append(generatedEdges, newEdge.ToGraphEdge())
	}

	// select 10 random users to add to the DA group. Starting Node is a Random user, Ending Node is the DA group
	eNode = generatedNodes[len(generatedNodes)-1].ID // -1 because the array starts at 0
	for i := 0; i < 10; i++ {
		x := r.Intn(nbUsers)
		sNode = generatedNodes[x].ID
		newEdge := generator.AddADEdge(r, "MemberOf", sNode, eNode) // r & 0 won't be used, they are place holder
		generatedEdges = append(generatedEdges, newEdge.ToGraphEdge())
	}

	// Generate a bunch of random edges

	for i := 0; i <= nbEdges; i++ {
		// We use `len(generatedNodes) -1` to calculate the nb of objects in our graph. This is to ensure we don't create Edges to objects that don't exist
		x := r.Intn(len(generatedNodes) - 1)
		sNode := generatedNodes[x]
		x = r.Intn(len(generatedNodes) - 1)
		eNode := generatedNodes[x]
		newEdge := generator.AddValidADEdge(r, "", sNode, eNode) // No edge type means it will be generated in the function.
		if newEdge.Kind != "" {                                  // If the function returned a valid edge for our objects, add it to the list of edges to generate. Otherwise select 2 new nodes
			generatedEdges = append(generatedEdges, newEdge.ToGraphEdge())
		} else {
			i--
		}
	}

	return generatedNodes, generatedEdges

}

func generateEntraGraph(r *rand.Rand) ([]generator.GenericIngestNode, []generator.GenericIngestEdge) {
	fmt.Printf("Entra is not yet supproted. Sorry :(")

	var generatedNodes []generator.GenericIngestNode
	var generatedEdges []generator.GenericIngestEdge

	return generatedNodes, generatedEdges

}

func main() {

	config.Validate()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var generatedNodes []generator.GenericIngestNode
	var generatedEdges []generator.GenericIngestEdge

	switch config.Type {
	case "AD":
		generatedNodes, generatedEdges = generateADGraph(r)
	case "Entra":
		generatedNodes, generatedEdges = generateEntraGraph(r)
	default:
		fmt.Printf("Could not find a method to build a graph that match you selected type.")
	}

	graph := generator.GenericIngestGraph{
		Graph: generator.GenericIngestGraphContent{
			Nodes: generatedNodes,
			Edges: generatedEdges,
		},
	}

	fileName := "test1.json"
	outputDir := "./output/test/"
	filePath := fmt.Sprintf(outputDir + fileName)
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(graph); err != nil {
		panic(err)
	}

	uploadFiles(outputDir)
}
