//  Copyright 2025 Specter Ops, Inc.
//
//  Licensed under the Apache License, Version 2.0
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//
//  SPDX-License-Identifier: Apache-2.0
//

package license

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	license "github.com/specterops/bloodhound/packages/go/stbernard/command/license/internal"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

const (
	Name  = "license"
	Usage = "Run license cmd to append license headers on bhce files"
)

type command struct {
	env environment.Environment
}

// Create a new instance of goimports command within the current environment
func Create(env environment.Environment) *command {
	return &command{
		env: env,
	}
}

// Usage of the command
func (s *command) Usage() string {
	return Usage
}

// Name of the command
func (s *command) Name() string {
	return Name
}

func (s *command) Parse(cmdIndex int) error {
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)
	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	return nil
}

func (s *command) Run() error {
	if err := license.Run(); err != nil {
		return fmt.Errorf("running license cmd: %w", err)
	}
	return nil
}
