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

package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type CustomNodeKinds []CustomNodeKind

func (s CustomNodeKinds) AuditData() AuditData {
	var data = make(AuditData)

	for i, kind := range s {
		data[fmt.Sprint(i)] = kind.AuditData()
	}

	return data
}

type CustomNodeKind struct {
	ID       int32                `json:"id"`
	KindName string               `json:"kindName"`
	Config   CustomNodeKindConfig `json:"config"`
}

func (s CustomNodeKind) AuditData() AuditData {
	return AuditData{
		"id":     s.ID,
		"kind":   s.KindName,
		"config": s.Config,
	}
}

type CustomNodeKindConfig struct {
	Icon CustomNodeIcon `json:"icon"`
}

type CustomNodeIcon struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func (s *CustomNodeKindConfig) Scan(value interface{}) error {
	if value == nil {
		*s = CustomNodeKindConfig{}
	}

	if bytes, ok := value.([]byte); !ok {
		return errors.New("type assertion to []byte failed for CustomNodeKindConfig")
	} else {
		return json.Unmarshal(bytes, &s)
	}
}

func (s CustomNodeKindConfig) Value() (driver.Value, error) {
	return json.Marshal(s)
}
