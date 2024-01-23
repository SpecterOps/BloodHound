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
	"github.com/specterops/bloodhound/packages/go/stbernard/command/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/modsync"
)

// Command enum represents our subcommands
type Command int

const (
	InvalidCommand Command = iota - 1
	ModSync
	EnvDump
	Analysis
)

// String implements Stringer for the Command enum
func (s Command) String() string {
	switch s {
	case ModSync:
		return modsync.Name
	case EnvDump:
		return envdump.Name
	case Analysis:
		return analysis.Name
	default:
		return "invalid command"
	}
}

// Commands returns our valid set of Command options
func Commands() []Command {
	return []Command{ModSync, EnvDump, Analysis}
}

// Commands usage returns a slice of Command usage statements indexed by their enum
func CommandsUsage() []string {
	var usage = make([]string, len(Commands()))

	usage[ModSync] = modsync.Usage
	usage[EnvDump] = envdump.Usage
	usage[Analysis] = analysis.Usage

	return usage
}
