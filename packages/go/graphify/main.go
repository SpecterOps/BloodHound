// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"log"

	"github.com/specterops/bloodhound/packages/go/graphify/graph"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

func main() {
	env := environment.NewEnvironment()

	gs, err := graph.NewCommunityGraphService()

	if err != nil {
		log.Fatal("Failed to create graph service: ", err)
	}

	command := graph.Create(env, gs)

	if err := command.Parse(); err != nil {
		log.Fatal("Failed to parse CLI args: ", err)
	} else if err := command.Run(); err != nil {
		log.Fatal("Failed to run command: ", err)
	}
}
