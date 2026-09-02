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

package harnesses

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

type NodeStyle struct {
	NodeColor string `json:"node-color"`
}

type RelationshipStyle struct {
	NodeColor string `json:"arrow-color"`
}

type Node struct {
	ID         string            `json:"id"`
	Caption    string            `json:"caption"`
	Labels     []string          `json:"labels"`
	Properties map[string]string `json:"properties"`
	Style      NodeStyle
}

type Relationship struct {
	ID         string            `json:"id"`
	From       string            `json:"fromId"`
	To         string            `json:"toId"`
	Kind       string            `json:"type"`
	Properties map[string]string `json:"properties"`
	Style      RelationshipStyle
}

type HarnessData struct {
	Nodes         []Node
	Relationships []Relationship
}

func filename() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("unable to get the current filename")
	}
	return filename, nil
}

func dir() (string, error) {
	filename, err := filename()
	if err != nil {
		return "", err
	}
	return filepath.Dir(filename), nil
}

func ReadHarness(harnessName string) (HarnessData, error) {
	if dir, err := dir(); err != nil {
		return HarnessData{}, err
	} else if jsonFile, err := os.Open(path.Join(dir, harnessName+".json")); err != nil {
		return HarnessData{}, err
	} else {
		defer jsonFile.Close()

		if byteValue, err := io.ReadAll(jsonFile); err != nil {
			return HarnessData{}, err
		} else {
			var result HarnessData
			json.Unmarshal(byteValue, &result)
			return result, nil
		}
	}
}
