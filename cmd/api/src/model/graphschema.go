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

package model

type GraphSchemaExtension struct {
	Serial

	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name"`
	Version     string `json:"version" validate:"required"`
	IsBuiltin   bool   `json:"is_builtin"`
}

func (GraphSchemaExtension) TableName() string {
	return "extensions"
}

func (s GraphSchemaExtension) AuditData() AuditData {
	return AuditData{
		"id":           s.ID,
		"name":         s.Name,
		"display_name": s.DisplayName,
		"version":      s.Version,
		"is_builtin":   s.IsBuiltin,
	}
}

type GraphSchemaExtensionProperty struct {
	Serial

	ExtensionID int32  `json:"extension_id"`
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name"`
	DataType    string `json:"data_type" validate:"required"`
	Description string `json:"description"`
}

func (GraphSchemaExtensionProperty) TableName() string {
	return "extension_schema_properties"
}

func (s GraphSchemaExtensionProperty) AuditData() AuditData {
	return AuditData{
		"id":           s.ID,
		"extension_id": s.ExtensionID,
		"name":         s.Name,
		"display_name": s.DisplayName,
		"data_type":    s.DataType,
		"description":  s.Description,
	}
}
