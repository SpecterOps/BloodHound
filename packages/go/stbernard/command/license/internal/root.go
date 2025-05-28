// Copyright 2025 Specter Ops, Inc.
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

package license

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func CheckRootFiles(wd string) error {
	rootFiles := []string{"LICENSE", "LICENSE.header"}

	for _, file := range rootFiles {
		if err := createLicenseFiles(wd + "/" + file); err != nil {
			return err
		}
	}
	return nil
}

// createLicenseFile validates and writes root license and license.header
func createLicenseFiles(path string) error {
	fmt.Printf("checking: %s\n", path)

	// check if the file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			s := strings.Contains(path, "header")
			switch s {
			case true:
				formattedHeader := GenerateLicenseHeader("")
				if err := os.WriteFile(path, []byte(strings.Join(formattedHeader, "")), 0777); err != nil {
					log.Fatal(err)
				}
				fmt.Printf("creating %v\n", path)
			default:
				if err := os.WriteFile(path, []byte(licenseContent), 0777); err != nil {
					return err
				}
				fmt.Printf("creating %v\n", path)
			}
		}
	} else {
		fmt.Printf("file: %s exists\n", path)
	}

	return nil
}
