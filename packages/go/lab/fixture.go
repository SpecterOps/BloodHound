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

	"github.com/specterops/bloodhound/lab/internal"
)

type depender interface {
	Dependencies() internal.Set[depender]
}

func NewFixture[T any](setup func(*Harness) (T, error), teardown func(*Harness, T) error) *Fixture[T] {
	return &Fixture[T]{
		dependencies: make(internal.Set[depender]),
		setup:        setup,
		teardown:     teardown,
	}
}

type Fixture[T any] struct {
	dependencies internal.Set[depender]
	setup        func(*Harness) (T, error)
	teardown     func(*Harness, T) error
}

func (s *Fixture[T]) Setup(harness *Harness) (T, error) {
	var out T
	if s.setup == nil {
		return out, fmt.Errorf("%T requires a non-nil setup function", s)
	} else if value, err := s.setup(harness); err != nil {
		return out, fmt.Errorf("failed to setup %T: %w", s, err)
	} else {
		return value, nil
	}
}

func (s *Fixture[T]) Teardown(harness *Harness, t T) error {
	if s.teardown != nil {
		return s.teardown(harness, t)
	}
	return nil
}

func (s *Fixture[T]) Dependencies() internal.Set[depender] {
	return s.dependencies
}

func hasCycle(consumer depender, producer depender) bool {
	var (
		visited                = make(internal.Set[depender])
		producerTransitiveDeps = []depender{producer}
	)
	for len(producerTransitiveDeps) > 0 {
		fixture := producerTransitiveDeps[len(producerTransitiveDeps)-1]
		producerTransitiveDeps = producerTransitiveDeps[:len(producerTransitiveDeps)-1]

		if fixture == consumer {
			return true
		}

		visited.Add(fixture)

		for dependency := range fixture.Dependencies() {
			if !visited.Has(dependency) {
				producerTransitiveDeps = append(producerTransitiveDeps, dependency)
			}
		}
	}
	return false
}

func setDependency(consumer depender, provider depender) error {
	if hasCycle(consumer, provider) {
		return fmt.Errorf("unable to set dependency for %T -> %T: cycle detected", consumer, provider)
	} else {
		consumer.Dependencies().Add(provider)
		return nil
	}
}

func SetDependency(consumer depender, provider depender) error {
	return setDependency(consumer, provider)
}

func UseFixture[Ctx TBConstraint, T any](ctx Ctx, fixture *Fixture[T]) (Assertions, T) {
	assert, harness := UseHarness(ctx, Pack(NewHarness(), fixture))
	value, ok := Unpack(harness, fixture)
	assert.True(ok, fmt.Sprintf("unable to unpack fixture %T", fixture))
	return assert, value
}
