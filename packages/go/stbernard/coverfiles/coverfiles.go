// Copyright 2024 Specter Ops, Inc.
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

package coverfiles

import (
	"errors"
	"fmt"
	"io"

	"golang.org/x/tools/cover"
)

var (
	// ErrNoProfiles tells the caller that no profiles were provided
	ErrNoProfiles = errors.New("no profiles")
)

// WriteProfile combines a list of cover profiles and writes them all to a single new file
func WriteProfile(w io.Writer, profiles []*cover.Profile) error {
	if len(profiles) == 0 {
		return ErrNoProfiles
	}

	if _, err := fmt.Fprintf(w, "mode: %s\n", profiles[0].Mode); err != nil {
		return err
	}

	for _, p := range profiles {
		for _, b := range p.Blocks {
			if _, err := fmt.Fprintf(w, "%s:%d.%d,%d.%d %d %d\n", p.FileName, b.StartLine, b.StartCol, b.EndLine, b.EndCol, b.NumStmt, b.Count); err != nil {
				return err
			}
		}
	}

	return nil
}
