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

package translate

import (
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func (s *Translator) buildProjection(scope *Scope) error {
	var (
		singlePartQuerySelect = pgsql.Select{}
	)

	singlePartQuerySelect.From = []pgsql.FromClause{{
		Source: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{scope.CurrentFrameBinding().Identifier},
		},
	}}

	if projectionConstraint, err := s.treeTranslator.ConsumeAll(); err != nil {
		return err
	} else if projection, err := buildExternalProjection(scope, s.projections.Projections); err != nil {
		return err
	} else {
		singlePartQuerySelect.Projection = projection
		singlePartQuerySelect.Where = projectionConstraint.Expression
	}

	s.query.Model.Body = singlePartQuerySelect

	if s.query.Skip.Set {
		s.query.Model.Offset = s.query.Skip
	}

	if s.query.Limit.Set {
		s.query.Model.Limit = s.query.Limit
	}

	if len(s.query.OrderBy) > 0 {
		s.query.Model.OrderBy = s.query.OrderBy
	}

	return nil
}

func (s *Translator) buildMatch(scope *Scope) error {
	for _, part := range s.match.Pattern.Parts {
		// Pattern can't be in scope at time of select as the pattern's scope directly depends on the
		// pattern parts
		if err := s.buildPatternPart(scope, part); err != nil {
			return err
		}

		// Declare the pattern variable in scope if set
		if part.PatternBinding.Set {
			scope.Declare(part.PatternBinding.Value.Identifier)
		}
	}

	return nil
}
