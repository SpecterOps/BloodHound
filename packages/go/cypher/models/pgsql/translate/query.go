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

package translate

import (
	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func (s *Translator) previousValidFrame(partFrame *Frame) (*Frame, bool) {
	if partFrame.Previous == nil {
		return nil, false
	}

	if currentQueryPart := s.query.CurrentPart(); currentQueryPart.Frame != nil && partFrame.Previous.Binding.Identifier == currentQueryPart.Frame.Binding.Identifier {
		// If the part's previous frame matches the query part's frame identifier then it's possible that
		// this current part is a multipart query part. In this case there still may be a valid frame
		// to source references from
		return currentQueryPart.Frame.Previous, currentQueryPart.Frame.Previous != nil
	}

	return partFrame.Previous, true
}

func (s *Translator) buildMultiPartSinglePartQuery(singlePartQuery *cypher.SinglePartQuery, cteChain []pgsql.CommonTableExpression) error {
	// Prepend the CTE chain to the model's
	currentPart := s.query.CurrentPart()
	currentPart.Model.CommonTableExpressions.Expressions = append(cteChain, currentPart.Model.CommonTableExpressions.Expressions...)

	return nil
}

func (s *Translator) buildSinglePartQuery(singlePartQuery *cypher.SinglePartQuery) error {
	if s.query.CurrentPart().HasDeletions() {
		if err := s.buildDeletions(s.scope); err != nil {
			s.SetError(err)
		}
	}

	// If there was no return specified end the CTE chain with a bare select
	if singlePartQuery.Return == nil {
		if literalReturn, err := pgsql.AsLiteral(1); err != nil {
			s.SetError(err)
		} else {
			s.query.CurrentPart().Model.Body = pgsql.Select{
				Projection: []pgsql.SelectItem{literalReturn},
			}
		}
	} else if err := s.buildTailProjection(); err != nil {
		s.SetError(err)
	}

	return nil
}

func (s *Translator) buildMultiPartQuery(singlePartQuery *cypher.SinglePartQuery) error {
	var multipartCTEChain []pgsql.CommonTableExpression

	// In order to author the multipart query part we first have to wrap it in a
	for _, part := range s.query.Parts[:len(s.query.Parts)-1] {
		// If the part has an empty inner CTE, make sure to remove it otherwise the keyword will still render
		if len(part.Model.CommonTableExpressions.Expressions) == 0 {
			part.Model.CommonTableExpressions = nil
		}

		// Autor the part as a nested CTE
		nextCTE := pgsql.CommonTableExpression{
			Query: *part.Model,
		}

		if part.Frame != nil {
			nextCTE.Alias = pgsql.TableAlias{
				Name: part.Frame.Binding.Identifier,
			}
		}

		if inlineSelect, err := s.buildInlineProjection(part); err != nil {
			return err
		} else {
			nextCTE.Query.Body = inlineSelect
		}

		multipartCTEChain = append(multipartCTEChain, nextCTE)
	}

	if err := s.buildMultiPartSinglePartQuery(singlePartQuery, multipartCTEChain); err != nil {
		return err
	}

	s.translation.Statement = *s.query.CurrentPart().Model
	return nil
}

func (s *Translator) translateMultiPartQueryPart() error {
	queryPart := s.query.CurrentPart()

	// Unwind nested frames
	return s.scope.UnwindToFrame(queryPart.Frame)
}

func (s *Translator) prepareSinglePartQueryPart(singlePartQuery *cypher.SinglePartQuery) error {
	s.query.AddPart(NewQueryPart(len(singlePartQuery.ReadingClauses), len(singlePartQuery.UpdatingClauses)))
	return nil
}

func (s *Translator) prepareMultiPartQueryPart(multiPartQueryPart *cypher.MultiPartQueryPart) error {
	newQueryPart := NewQueryPart(len(multiPartQueryPart.ReadingClauses), len(multiPartQueryPart.UpdatingClauses))

	// All multipart query parts must be wrapped in a nested CTE
	if mpFrame, err := s.scope.PushFrame(); err != nil {
		return err
	} else {
		newQueryPart.Frame = mpFrame
	}

	s.query.AddPart(newQueryPart)
	return nil
}
