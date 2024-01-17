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
)

type errorTransactionWrapper struct {
	err error
}

func newErrorTransactionWrapper(err error) errorTransactionWrapper {
	return errorTransactionWrapper{
		err: err,
	}
}

func (s errorTransactionWrapper) Run(cypher string, params map[string]any) (neo4j.Result, error) {
	return nil, s.err
}

func (s errorTransactionWrapper) Commit() error {
	return s.err
}

func (s errorTransactionWrapper) Rollback() error {
	return s.err
}

func (s errorTransactionWrapper) Close() error {
	return s.err
}
