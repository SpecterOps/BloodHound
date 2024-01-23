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

package envdump

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	Name  = "envdump"
	Usage = "Dump your environment variables"
)

type Config struct {
	Environment []string
}

type command struct {
	config Config
}

func (s command) Name() string {
	return Name
}

func (s command) Usage() string {
	return Usage
}

func (s command) Run() error {
	fmt.Print("Environment:\n\n")
	for _, env := range s.config.Environment {
		envTuple := strings.SplitN(env, "=", 2)
		fmt.Printf("%s: %s\n", envTuple[0], envTuple[1])
	}
	fmt.Print("\n")

	return nil
}

func Create(config Config) (command, error) {
	envdumpCmd := flag.NewFlagSet(Name, flag.ExitOnError)

	envdumpCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n", Usage, filepath.Base(os.Args[0]), Name)
	}

	if err := envdumpCmd.Parse(os.Args[2:]); err != nil {
		envdumpCmd.Usage()
		return command{}, fmt.Errorf("failed to parse %s command: %w", Name, err)
	} else {
		return command{config: config}, nil
	}
}
