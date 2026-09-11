// Copyright 2026 Specter Ops, Inc.
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

package services

import "errors"

// ErrExtensionNotFound indicates that no extension exists for the requested id.
var ErrExtensionNotFound = errors.New("extension not found")

// Extension is the domain representation of a schema_extensions row.
type Extension struct {
	ID          int32
	Name        string
	DisplayName string
	Namespace   string
	IsBuiltin   bool
	Version     string
}
