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

package command

import (
	"github.com/specterops/bloodhound/packages/go/stbernard/command/analysis"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/builder"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/generate"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/modsync"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/tester"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

type Command interface {
	// Name gets the name of the Command
	Name() string
	// Usage gets the usage string for the Command
	Usage() string
	// Parse parses flags for the command using the command index as the starting point
	Parse(cmdIdx int) error
	// Run will run the command and return any errors
	Run() error
}

// Commands returns our valid set of Command options
func Commands() []Command {
	var env = environment.NewEnvironment()

	return []Command{
		envdump.Create(env),
		modsync.Create(env),
		generate.Create(env),
		analysis.Create(env),
		tester.Create(env),
		builder.Create(env),
	}
}
