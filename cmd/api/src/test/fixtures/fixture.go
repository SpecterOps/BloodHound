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

package fixtures

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"path/filepath"

	"github.com/specterops/bloodhound/src/test"
)

//go:embed fixtures
var fixtures embed.FS

type ErrorHandler func(path string, err error)

func NewTestErrorHandler(testRef test.Controller) ErrorHandler {
	return func(path string, err error) {
		testRef.Fatalf("Failed loading fixture %s: %v", path, err)
	}
}

func NewPanicErrorHandler() ErrorHandler {
	return func(path string, err error) {
		panic(fmt.Sprintf("Failed loading fixture %s: %v", path, err))
	}
}

type Loader interface {
	Get(path string) []byte
	GetString(path string) string
	GetReader(path string) io.ReadCloser
}

type loader struct {
	errorHandler ErrorHandler
}

func NewLoader(errorHandler ErrorHandler) Loader {
	return loader{
		errorHandler: errorHandler,
	}
}

func (s loader) Get(path string) []byte {
	if content, err := fixtures.ReadFile(filepath.Join("fixtures", path)); err != nil {
		s.errorHandler(path, err)
		return nil
	} else {
		content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))
		return content
	}
}

func (s loader) GetString(path string) string {
	return string(s.Get(path))
}

func (s loader) GetReader(path string) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(s.Get(path)))
}
