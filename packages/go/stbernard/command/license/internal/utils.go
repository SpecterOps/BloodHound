// Copyright Specter Ops, Inc.
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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func parseFileExtension(path string) string {
	return filepath.Ext(path)
}

func generateLicenseHeader(commentPrefix string) []string {
	if commentPrefix == "" {
		return nil
	}

	var formattedHeader []string
	s := strings.Split(licenseHeader, "\n")

	for _, line := range s {
		f := fmt.Sprintf("%v %v\n", commentPrefix, line)
		formattedHeader = append(formattedHeader, f)
	}

	year := getCurrentYear()

	// Find and replace the copyright line instead of using hardcoded index
	for i, line := range formattedHeader {
		if strings.HasPrefix(line, "Copyright") {
			formattedHeader[i] = fmt.Sprintf("Copyright %s Specter Ops, Inc.", year)
			break
		}
	}
	return formattedHeader
}

func writeFile(path string, formattedHeaderContent []string) error {
	// Get original file info to preserve permissions
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	originalPerm := fileInfo.Mode().Perm()

	var newContent []string
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent = append(formattedHeaderContent, string(data))

	if err := os.WriteFile(path, []byte(strings.Join(newContent, "")), originalPerm); err != nil {
		return err
	}
	return nil
}

func generateXMLLicenseHeader() []string {
	s := fmt.Sprintf("<!--%v\n-->", licenseHeader)

	formattedHeaderContent := strings.Split(s, "\n")

	year := getCurrentYear()
	// Find and replace the copyright line instead of using hardcoded index
	for i, line := range formattedHeaderContent {
		if strings.HasPrefix(line, "Copyright") {
			formattedHeaderContent[i] = fmt.Sprintf("Copyright %s Specter Ops, Inc.", year)
			break
		}
	}

	return formattedHeaderContent
}

func writeXMLFile(path string, formattedHeaderContent []string) error {
	// Get original file info to preserve permissions
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	originalPerm := fileInfo.Mode().Perm()

	var newContent []string
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent = append(formattedHeaderContent, string(data))

	if err := os.WriteFile(path, []byte(strings.Join(newContent, "\n")), originalPerm); err != nil {
		return err
	}
	return nil
}

func getCurrentYear() string {
	year := time.Now().Year()
	return strconv.Itoa(year)
}
