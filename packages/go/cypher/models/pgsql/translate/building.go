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

func (s *Translator) buildInlineProjection(part *QueryPart) (pgsql.Select, error) {
	var sqlSelect pgsql.Select

	if part.projections.Frame != nil {
		sqlSelect.From = []pgsql.FromClause{{
			Source: part.projections.Frame.Binding.Identifier,
		}}
	}

	if projectionConstraint, err := s.treeTranslator.ConsumeAll(); err != nil {
		return sqlSelect, err
	} else {
		sqlSelect.Where = projectionConstraint.Expression
	}

	for _, projection := range part.projections.Items {
		builtProjection := projection.SelectItem

		if projection.Alias.Set {
			builtProjection = &pgsql.AliasedExpression{
				Expression: builtProjection,
				Alias:      projection.Alias,
			}
		}

		sqlSelect.Projection = append(sqlSelect.Projection, builtProjection)
	}

	if len(part.projections.GroupBy) > 0 {
		for _, groupBy := range part.projections.GroupBy {
			sqlSelect.GroupBy = append(sqlSelect.GroupBy, groupBy)
		}
	}

	return sqlSelect, nil
}

func (s *Translator) buildTailProjection() error {
	var (
		currentPart           = s.query.CurrentPart()
		currentFrame          = s.query.Scope.CurrentFrame()
		singlePartQuerySelect = pgsql.Select{}
	)

	singlePartQuerySelect.From = []pgsql.FromClause{{
		Source: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{currentFrame.Binding.Identifier},
		},
	}}

	if projectionConstraint, err := s.treeTranslator.ConsumeAll(); err != nil {
		return err
	} else if projection, err := buildExternalProjection(s.query.Scope, currentPart.projections.Items); err != nil {
		return err
	} else if err := RewriteExpressionIdentifiers(projectionConstraint.Expression, currentFrame.Binding.Identifier, nil); err != nil {
		return err
	} else {
		singlePartQuerySelect.Projection = projection
		singlePartQuerySelect.Where = projectionConstraint.Expression
	}

	currentPart.Model.Body = singlePartQuerySelect

	if currentPart.Skip.Set {
		currentPart.Model.Offset = currentPart.Skip
	}

	if currentPart.Limit.Set {
		currentPart.Model.Limit = currentPart.Limit
	}

	if len(currentPart.OrderBy) > 0 {
		currentPart.Model.OrderBy = currentPart.OrderBy
	}

	return nil
}

func (s *Translator) buildMatch() error {
	for _, part := range s.query.CurrentPart().match.Pattern.Parts {
		// Pattern can't be in scope at time of select as the pattern's scope directly depends on the
		// pattern parts
		if err := s.buildPatternPart(part); err != nil {
			return err
		}

		// Declare the pattern variable in scope if set
		if part.PatternBinding.Set {
			s.query.Scope.Declare(part.PatternBinding.Value.Identifier)
		}
	}

	return nil
}
