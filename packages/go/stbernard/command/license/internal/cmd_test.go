// Copyright 2026 Specter Ops, Inc.
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
	"path/filepath"
	"testing"
)

func TestIsIgnoredDir(t *testing.T) {
	testCases := []struct {
		name     string
		dirName  string
		expected bool
	}{
		{name: "storybook build output is ignored", dirName: "storybook-static", expected: true},
		{name: "node_modules is ignored", dirName: "node_modules", expected: true},
		{name: "dist is ignored", dirName: "dist", expected: true},
		{name: "git metadata is ignored", dirName: ".git", expected: true},
		{name: "source directory is not ignored", dirName: "src", expected: false},
		{name: "arbitrary directory is not ignored", dirName: "cmd", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if actual := isIgnoredDir(testCase.dirName); actual != testCase.expected {
				t.Errorf("isIgnoredDir(%q) = %v, want %v", testCase.dirName, actual, testCase.expected)
			}
		})
	}
}

func TestMatchesIgnorePath(t *testing.T) {
	testCases := []struct {
		name     string
		relPath  string
		expected bool
	}{
		{name: "exact ignored path", relPath: "justfile", expected: true},
		{name: "prefix of ignored path", relPath: filepath.Join("cmd", "ui", "playwright", "foo.ts"), expected: true},
		{name: "static assets are ignored", relPath: filepath.Join("cmd", "api", "src", "api", "static", "assets", "app.js"), expected: true},
		{name: "unrelated path is not ignored", relPath: filepath.Join("cmd", "api", "src", "main.go"), expected: false},
		{name: "empty path is not ignored", relPath: "", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if actual := matchesIgnorePath(testCase.relPath); actual != testCase.expected {
				t.Errorf("matchesIgnorePath(%q) = %v, want %v", testCase.relPath, actual, testCase.expected)
			}
		})
	}
}

func TestIsDisallowedExtension(t *testing.T) {
	testCases := []struct {
		name     string
		ext      string
		expected bool
	}{
		{name: "json is disallowed", ext: ".json", expected: true},
		{name: "markdown is disallowed", ext: ".md", expected: true},
		{name: "lock file is disallowed", ext: ".lock", expected: true},
		{name: "go source is allowed", ext: ".go", expected: false},
		{name: "typescript source is allowed", ext: ".ts", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if actual := isDisallowedExtension(testCase.ext); actual != testCase.expected {
				t.Errorf("isDisallowedExtension(%q) = %v, want %v", testCase.ext, actual, testCase.expected)
			}
		})
	}
}
