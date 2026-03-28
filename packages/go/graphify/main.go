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
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/packages/go/graphify/graph"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
)

func main() {
	env, err := environment.NewEnvironment()
	if err != nil {
		slog.Error("Failed to init environment", attr.Error(err))
		os.Exit(1)
	}

	gs, err := graph.NewCommunityGraphService()
	if err != nil {
		slog.Error("Failed to create graph service", attr.Error(err))
		os.Exit(1)
	}

	command := graph.Create(env, gs)

	if err := command.Parse(); err != nil {
		slog.Error("Failed to parse CLI args", attr.Error(err))
		os.Exit(1)
	} else if err := command.Run(); err != nil {
		slog.Error("Failed to run command", attr.Error(err))
		os.Exit(1)
	}
}
