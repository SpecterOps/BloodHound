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
	"bufio"
	"fmt"
	"os"
	"strings"
)

func dirCheck(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("failed getting file stat %v\n", err)
	}

	return fileInfo.IsDir()
}

func ignorePathValidation(ignorPaths []string, wd, path string) bool {
	for _, ig := range ignorPaths {
		if strings.Contains(path, ig) {
			return false
		}
	}

	return true
}

func isHeaderPresent(path string) (bool, error) {
	fileReader, err := os.Open(path)
	if err != nil {
		return false, err
	}

	// check for license header
	fileScanner := bufio.NewScanner(fileReader)
	fileScanner.Split(bufio.ScanLines)
	var lines []string
	linestoRead := 3
	linesRead := 0
	for fileScanner.Scan() && linesRead < linestoRead {
		if i := 0; i <= 2 {
			lines = append(lines, fileScanner.Text())
			linesRead++
		}
	}
	data := strings.Join(lines, "\n")
	return strings.Contains(data, "Copyright"), nil
}
