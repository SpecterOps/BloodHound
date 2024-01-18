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
	"errors"
	"log"
	"strings"

	"github.com/specterops/bloodhound/packages/go/stbernard/command"
)

func main() {
	if cmd, err := command.ParseCLI(); err != nil {
		if errors.Is(err, command.ErrNoCmd) {
			log.Fatal("No command specified")
		} else {
			log.Fatalf("Error while parsing command: %v", err)
		}
	} else if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run command %s: %v", cmd.Name(), err)
	} else {
		log.Printf("%s completed successfully", strings.ToUpper(cmd.Name()))
	}
}
