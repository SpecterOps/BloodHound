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

package neo4j

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type internalResult struct {
	query        string
	err          error
	driverResult neo4j.Result
}

func NewResult(query string, err error, driverResult neo4j.Result) graph.Result {
	return &internalResult{
		query:        query,
		err:          err,
		driverResult: driverResult,
	}
}

func (s *internalResult) Values() (graph.ValueMapper, error) {
	return NewValueMapper(s.driverResult.Record().Values), nil
}

func (s *internalResult) Scan(targets ...any) error {
	if values, err := s.Values(); err != nil {
		return err
	} else {
		return values.Scan(targets...)
	}
}

func (s *internalResult) Next() bool {
	return s.driverResult.Next()
}

func (s *internalResult) Error() error {
	if s.err != nil {
		return s.err
	}

	if s.driverResult != nil && s.driverResult.Err() != nil {
		return graph.NewError(s.query, s.driverResult.Err())
	}

	return nil
}

func (s *internalResult) Close() {
	if s.driverResult != nil {
		// Ignore the results of this call. This is called only as a best-effort attempt at a close
		s.driverResult.Consume()
	}
}
