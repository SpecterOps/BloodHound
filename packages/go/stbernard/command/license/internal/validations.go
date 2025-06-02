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
package license

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func isHeaderPresent(path string) (bool, error) {
	const linesToRead = 20

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// check for license header
	r := bufio.NewReader(file)

	for range linesToRead {
		if line, err := r.ReadString('\n'); !errors.Is(err, io.EOF) && err != nil {
			return false, fmt.Errorf("could not read line: %w", err)
		} else if strings.Contains(line, "SPDX-License-Identifier: Apache-2.0") {
			return true, nil
		}
	}

	return false, nil
}
