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
	"fmt"
	"log"
	"os"
	"path/filepath"

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
				fmt.Printf("File: %s\n", fullPath)
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

	user := generator.AddADUser()
	fmt.Printf("My user is: ", user)
	fmt.Printf("\n")
	computer := generator.AddADComputer()
	fmt.Printf("My computer is: ", computer)

	//dir := "./output/ad_sample"
	//uploadFiles(dir)
}
