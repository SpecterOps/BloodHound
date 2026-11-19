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

import (
	"context"
	"errors"
	"time"

	"github.com/specterops/bloodhound/packages/go/params"
)

var (
	// ErrExtensionNotFound indicates that no extension exists for the requested id.
	ErrExtensionNotFound = errors.New("extension not found")

	ErrDuplicateExtensionNamespace = errors.New("duplicate extension namespace")

	ErrDuplicateExtensionName = errors.New("duplicate extension name")

	ErrGraphExtensionBuiltIn = errors.New("cannot modify a built-in graph extension")
)

// Extension is the domain representation of a schema_extensions row.
type Extension struct {
	ID          int32
	Name        string
	DisplayName string
	Namespace   string
	IsBuiltin   bool
	Version     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *Service) GetExtension(ctx context.Context, id int32) (Extension, error) {
	return s.db.GetExtension(ctx, id)
}

func (s *Service) ListExtensions(ctx context.Context, extensionFilters params.Filters, sort params.SortItems, skip, limit int) ([]Extension, int, error) {
	return s.db.ListExtensions(ctx, extensionFilters, sort, skip, limit)
}
