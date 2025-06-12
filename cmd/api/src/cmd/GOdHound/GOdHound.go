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
					fmt.Printf("File: %v uploaded", fin)
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

func main() {

	config.Validate()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	nbUsers := 100
	nbComps := 100
	//nbGroups := nbUsers / 10
	nbEdges := 25 //nbUsers * nbComps * nbGroups / 3

	//var generatedUsers []generator.ADUser
	var generatedNodes []generator.GenericIngestNode

	for i := 0; i <= nbUsers; i++ {
		newUser := generator.AddADUser(r, config.Domain.Name, int32(i))
		generatedNodes = append(generatedNodes, newUser.ToGraphNode())
	}

	//var generatedComps []generator.ADComputer
	for i := 1; i <= nbComps; i++ {
		y := nbUsers + i
		newComp := generator.AddADComputer(r, config.Domain.Name, int32(y))
		generatedNodes = append(generatedNodes, newComp.ToGraphNode())
	}

	//var jsonNodes generator.GenericIngestNode
	//jsonNodes := generator.ToGraphNode()

	var generatedEdges []generator.GenericIngestEdge
	for i := 0; i <= nbEdges; i++ {
		newEdge := generator.AddADEdge(nbUsers + nbComps)
		generatedEdges = append(generatedEdges, newEdge.ToGraphEdge())
	}

	graph := generator.GenericIngestGraph{
		Graph: generator.GenericIngestGraphContent{
			Nodes: generatedNodes,
			Edges: generatedEdges,
		},
	}

	fileName := "test1.json"
	/*
		generator.CreateHeader(file)
		generator.AddNodes(file)
		generator.AddEdges(file)
		generator.CreateFooter(file)

		fmt.Printf("My user is: %+v\n", generatedUsers)
		fmt.Printf("\n\n")
		computer := generator.AddADComputer()
		fmt.Printf("My computer is: ", computer)
	*/
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
