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

package analyzer

import (
	"errors"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type Analyzer struct {
	handlers []func(stack *model.WalkStack, node model.Expression) error
}

func (s *Analyzer) walkFunc(stack *model.WalkStack, expression model.Expression) error {
	var errs []error

	for _, handler := range s.handlers {
		if err := handler(stack, expression); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (s *Analyzer) analyze(query any, extensions ...model.CollectorFunc) error {
	return model.Walk(query, model.NewVisitor(s.walkFunc, nil), extensions...)
}

func Analyze(query any, registrationFunc func(analyzerInst *Analyzer), extensions ...model.CollectorFunc) error {
	analyzer := &Analyzer{}
	registrationFunc(analyzer)

	return analyzer.analyze(query, extensions...)
}

type typedVisitor[T model.Expression] func(stack *model.WalkStack, node T) error

func WithVisitor[T model.Expression](analyzer *Analyzer, visitorFunc typedVisitor[T]) {
	analyzer.handlers = append(analyzer.handlers, func(walkStack *model.WalkStack, node model.Expression) error {
		if typedNode, typeOK := node.(T); typeOK {
			if err := visitorFunc(walkStack, typedNode); err != nil {
				return err
			}
		}

		return nil
	})
}

func QueryComplexity(query *model.RegularQuery) (*ComplexityMeasure, error) {
	var (
		analyzer = &Analyzer{}
		measure  = &ComplexityMeasure{
			nodeLookupKinds: map[string]graph.Kinds{},
		}
	)

	WithVisitor(analyzer, measure.onPatternPart)
	WithVisitor(analyzer, measure.onNodePattern)
	WithVisitor(analyzer, measure.onProjection)
	WithVisitor(analyzer, measure.onRelationshipPattern)
	WithVisitor(analyzer, measure.onFunctionInvocation)
	WithVisitor(analyzer, measure.onKindMatcher)
	WithVisitor(analyzer, measure.onQuantifier)
	WithVisitor(analyzer, measure.onSortItem)
	WithVisitor(analyzer, measure.onPartialComparison)
	WithVisitor(analyzer, measure.onWhere)

	// Mutations
	WithVisitor(analyzer, measure.onCreate)
	WithVisitor(analyzer, measure.onDelete)
	WithVisitor(analyzer, measure.onMerge)
	WithVisitor(analyzer, measure.onRemove)
	WithVisitor(analyzer, measure.onSet)

	if err := analyzer.analyze(query); err != nil {
		return nil, err
	}

	// Wrap up with a call to the exit function
	measure.onExit()

	return measure, nil
}
