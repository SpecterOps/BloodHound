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
	"github.com/specterops/bloodhound/src/cmd/GOdHound/client"
	"os"
)

func fatal(value any) {
	fmt.Printf("%v", value)
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fatal(fmt.Sprintf(format, args...))
}

/*
Example Ingest File Format for Generic Ingest

{
  "graph": {
    "nodes": [
      {
        "kinds": [
          "User",
          "Base"
        ],
        "id": "1234-4567-13456-5120",
        "properties": {
          "name": "A User",
          "objectid": "1234-4567-13456-5120"
        }
      },
      {
        "kinds": [
          "Group",
          "Base"
        ],
        "id": "1234-4567-13456-5121",
        "properties": {
          "name": "A Group",
          "objectid": "1234-4567-13456-5121"
        }
      }
    ],
    "edges": [
      {
        "kind": "MemberOf",
        "start": {
          "match_by": "id",
          "value": "1234-4567-13456-5120",
          "kind": "Base"
        },
        "end": {
          "match_by": "id",
          "value": "1234-4567-13456-5121",
          "kind": "Base"
        },
        "properties": {
          "thing": "that thing I sent you"
        }
      }
    ]
  }
}
*/

func main() {
	var (
		token = &client.TokenCredentialsHandler{
			TokenID:  "ffa20bcd-5600-4e68-9c8e-49b5a0186a90",
			TokenKey: "iGd9tuDAEYEWp6YGyiYV6/e+mt1QmuV9S2s8ld///IMQyKsvmRcTLw==",
		}
	)

	if apiClient, err := client.NewClient("http://localhost:8080", token); err != nil {
		fatalf("Failed to create client: %v", err)
	} else {
		if version, err := apiClient.Version(); err != nil {
			fatalf("Failed to get version: %v", err)
		} else {
			fmt.Printf("Version: %s\n", version)
		}

		if newFileUploadJob, err := apiClient.StartFileUploadJob(); err != nil {
			fatal(err)
		} else {
			if fin, err := os.Open("/home/zinic/test.json"); err != nil {
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
		}
	}
}
