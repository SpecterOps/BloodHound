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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func parseFileExtension(path string) string {
	ext := ""
	parts := strings.Split(path, ".")

	if len(parts) >= 2 { // add support for files with multiple dots
		ext = fmt.Sprintf(".%s", parts[len(parts)-1])
	} else {
		ext = filepath.Ext(path)
	}

	return ext
}

func generateLicenseHeader(commentPrefix string) []string {
	var formattedHeader []string
	s := strings.Split(licenseHeader, "\n")

	for _, line := range s {
		f := fmt.Sprintf("%v  %v\n", commentPrefix, line)
		formattedHeader = append(formattedHeader, f)
	}

	year := getCurrentYear()
	formattedHeader[0] = fmt.Sprintf("%s  Copyright %s Specter Ops, Inc.\n", commentPrefix, year)

	return formattedHeader
}

func writeFile(path string, formattedHeaderContent []string) error {
	var newContent []string
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	formattedHeaderContent = append(formattedHeaderContent, "\n")
	newContent = append(formattedHeaderContent, string(data))

	if err := os.WriteFile(path, []byte(strings.Join(newContent, "")), 0666); err != nil {
		return err
	}
	return nil
}

func generateXMLLicenseHeader() []string {
	s := fmt.Sprintf("<!-- %v \n-->", licenseHeader)

	formattedHeaderContent := strings.Split(s, "\n")

	year := getCurrentYear()
	formattedHeaderContent[0] = fmt.Sprintf("<!-- \nCopyright %s Specter Ops, Inc.", year)

	return formattedHeaderContent
}

func writeXMLFile(path string, formattedHeaderContent []string) error {

	var newContent []string
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	formattedHeaderContent = append(formattedHeaderContent, "\n")
	newContent = append(formattedHeaderContent, string(data))

	if err := os.WriteFile(path, []byte(strings.Join(newContent, "\n")), 0666); err != nil {
		return err
	}
	return nil
}

func getCurrentYear() string {
	year := time.Now().Year()
	return strconv.Itoa(year)
}
