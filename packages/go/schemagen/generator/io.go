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

package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	defaultPackageDirPermission os.FileMode = 0755
	defaultSourceFilePermission os.FileMode = 0644

	fileOpenMode = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
)

type Writable interface {
	Render(writer io.Writer) error
}

func WriteSourceFile(output Writable, path string) error {
	dirPath := filepath.Dir(path)

	// Ensure the directory exists and if not attempt to create it and its dependencies
	if _, err := os.Stat(dirPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(dirPath, defaultPackageDirPermission); err != nil {
			return err
		}
	}

	if fout, err := os.OpenFile(path, fileOpenMode, defaultSourceFilePermission); err != nil {
		return err
	} else {
		defer fout.Close()
		return output.Render(fout)
	}
}

func FindGolangWorkspaceRoot() (string, error) {
	if workingDir, err := os.Getwd(); err != nil {
		return "", err
	} else {
		cursor := workingDir

		for cursor != "" && cursor != "/" {
			if fileList, err := os.ReadDir(cursor); err != nil {
				return "", err
			} else {
				for _, fileInfo := range fileList {
					if fileInfo.Name() == "go.work" {
						return cursor, nil
					}
				}
			}

			cursor = filepath.Dir(cursor)
		}

		if cursor == "" || cursor == "/" {
			return "", fmt.Errorf("unable to find project root from working directory path %s", workingDir)
		}

		return cursor, nil
	}
}
