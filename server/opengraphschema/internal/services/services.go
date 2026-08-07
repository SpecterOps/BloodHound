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

package services

//go:generate go tool mockery

import (
	"context"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
)

// Database describes the persistence capabilities the opengraphschema Service
// requires. Only the environment lookup used to resolve environment kinds is
// exercised by the consumers of this slice.
type Database interface {
	GetEnvironmentsFiltered(ctx context.Context, onlyBuiltin bool) ([]model.SchemaEnvironment, error)
}

// Service implements the opengraphschema use cases on top of a Database
// implementation.
type Service struct {
	db Database
}

// NewService constructs a Service backed by the supplied Database implementation.
func NewService(databaseInterface Database) *Service {
	return &Service{db: databaseInterface}
}

// GetEnvironmentKindsAndSchemaEnvironmentData returns all environment kinds as
// graph.Kinds and a map of their schema environments. When onlyBuiltin is true
// only builtin environment kinds are returned.
func (s *Service) GetEnvironmentKindsAndSchemaEnvironmentData(ctx context.Context, onlyBuiltin bool) (graph.Kinds, model.EnvironmentKindsToEnvironment, error) {
	if environments, err := s.db.GetEnvironmentsFiltered(ctx, onlyBuiltin); err != nil {
		return nil, nil, err
	} else {
		var (
			environmentKinds     = make([]graph.Kind, 0, len(environments))
			envKindToEnvironment = make(model.EnvironmentKindsToEnvironment, len(environments))
		)

		for _, env := range environments {
			environmentKinds = append(environmentKinds, graph.StringKind(env.EnvironmentKindName))
			envKindToEnvironment[env.EnvironmentKindName] = env
		}

		return environmentKinds, envKindToEnvironment, nil
	}
}
