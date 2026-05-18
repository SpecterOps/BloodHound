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

package analysis

import (
	"context"
	"errors"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// ErrNotFound is returned by Store operations when no matching analysis request row exists.
var ErrNotFound = errors.New("analysis request not found")

// Store adapts a lower-level analysis request persistence implementation so that callers
// only need to interact with appdb-level sentinels rather than database driver errors.
type Store struct {
	persistence database.AnalysisRequestData
}

// NewStore returns a Store backed by the provided persistence implementation.
func NewStore(persistence database.AnalysisRequestData) *Store {
	return &Store{persistence: persistence}
}

// GetAnalysisRequest returns the currently pending analysis request, or ErrNotFound when
// no request is present.
func (s *Store) GetAnalysisRequest(ctx context.Context) (model.AnalysisRequest, error) {
	var (
		analysisRequest model.AnalysisRequest
		err             error
	)

	analysisRequest, err = s.persistence.GetAnalysisRequest(ctx)
	if errors.Is(err, database.ErrNotFound) {
		return analysisRequest, ErrNotFound
	}
	return analysisRequest, err
}
