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

package lab

import (
	"fmt"
	"reflect"

	"github.com/specterops/bloodhound/lab/internal"
)

func NewHarness() *Harness {
	return &Harness{
		fixtures: make(map[any]any),
	}
}

type Harness struct {
	fixtures internal.Map[any, any]
}

func (s *Harness) Setup() (*Harness, error) {
	topoSort := s.topologicalSort()
	logTopology(topoSort)
	for i := len(topoSort) - 1; i >= 0; i-- {
		fixture := topoSort[i]
		if result := invoke(fixture, "Setup", s); len(result) != 2 {
			return s, fmt.Errorf("unable to setup %T: invalid setup method", fixture)
		} else if !result[1].IsNil() {
			return s, result[1].Interface().(error)
		} else {
			s.fixtures[fixture] = result[0].Interface()
		}
	}

	return s, nil
}

func (s *Harness) Teardown(harness *Harness) error {
	for _, fixture := range s.topologicalSort() {
		if result := invoke(fixture, "Teardown", harness, s.fixtures[fixture]); len(result) != 1 {
			return fmt.Errorf("unable to teardown %T: invalid teardown method", fixture)
		} else if !result[0].IsNil() {
			return result[0].Interface().(error)
		}
	}
	return nil
}

func (s Harness) topologicalSort() []any {
	var (
		inDegree  = make(map[any]int)
		sorted    = make([]any, 0)
		unvisited = make([]any, 0)
	)

	for fixture := range s.fixtures {
		for dependency := range fixture.(depender).Dependencies() {
			inDegree[dependency]++
		}
	}

	for fixture := range s.fixtures {
		if inDegree[fixture] == 0 {
			unvisited = append(unvisited, fixture)
		}
	}

	for len(unvisited) > 0 {
		fixture := unvisited[0]
		unvisited = unvisited[1:]
		sorted = append(sorted, fixture)

		for dependency := range fixture.(depender).Dependencies() {
			inDegree[dependency]--
			if inDegree[dependency] == 0 {
				unvisited = append(unvisited, dependency)
			}
		}
	}

	return sorted
}

func invoke(object any, method string, params ...any) []reflect.Value {
	rMethod := reflect.ValueOf(object).MethodByName(method)
	args := make([]reflect.Value, len(params))
	for i, param := range params {
		if param == nil {
			args[i] = reflect.Zero(rMethod.Type().In(1))
		} else {
			args[i] = reflect.ValueOf(param)
		}
	}
	return rMethod.Call(args)
}

func Pack[T any](harness *Harness, fixture *Fixture[T]) *Harness {
	return pack(harness, fixture)
}

func pack(harness *Harness, fixture any) *Harness {
	if !harness.fixtures.Has(fixture) {
		harness.fixtures.Set(fixture, nil)

		for dependency := range fixture.(depender).Dependencies() {
			pack(harness, dependency)
		}
	}
	return harness
}

func Unpack[T any](harness *Harness, fixture *Fixture[T]) (T, bool) {
	var empty T
	if value, ok := harness.fixtures[fixture]; !ok {
		return empty, false
	} else if out, ok := value.(T); !ok {
		return empty, false
	} else {
		return out, ok
	}
}

func UseHarness[Ctx TBConstraint](ctx Ctx, harness *Harness) (Assertions, *Harness) {
	return UseHooks(ctx, harness.Setup, harness.Teardown)
}
