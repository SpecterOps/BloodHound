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

package util

import (
	"errors"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"strings"
	"sync"
)

type ErrorCollector interface {
	Add(err error)
	Combined() error
}

type errorCollector struct {
	errors []error
	lock   *sync.Mutex
}

func NewErrorCollector() ErrorCollector {
	return &errorCollector{
		lock: &sync.Mutex{},
	}
}

func (s *errorCollector) Add(err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.errors = append(s.errors, err)
}

func (s *errorCollector) Combined() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.errors) > 0 {
		return errors.Join(s.errors...)
	}

	return nil
}

func IsNeoTimeoutError(err error) bool {
	if castError, ok := err.(*neo4j.Neo4jError); !ok {
		return false
	} else {
		return strings.Contains(castError.Code, "TransactionTimedOut")
	}
}
