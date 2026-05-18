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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	serverAppdbAnalysis "github.com/specterops/bloodhound/server/appdb/analysis"
)

// ErrNoPendingRequest indicates that there is no analysis request currently pending.
var ErrNoPendingRequest = errors.New("no pending analysis request")

// Database describes the persistence capabilities the analysis Service requires. Implementations
// are expected to translate driver- or ORM-specific not-found errors into appdb-level sentinels
// so that the Service can map them to its own failure-mode errors.
type Database interface {
	GetAnalysisRequest(ctx context.Context) (model.AnalysisRequest, error)
}

// Service implements the analysis use cases on top of a Database implementation.
type Service struct {
	db Database
}

// NewService constructs a Service backed by the supplied Database implementation.
func NewService(databaseInterface Database) *Service {
	return &Service{db: databaseInterface}
}

// GetRequest returns the currently pedatabaseInterfacending analysis request. ErrNoPendingRequest is returned
// when no request is pending; any other error indicates a failure servicing the request.
func (s *Service) GetRequest(ctx context.Context) (model.AnalysisRequest, error) {
	var (
		analysisRequest model.AnalysisRequest
		err             error
	)

	analysisRequest, err = s.db.GetAnalysisRequest(ctx)
	if errors.Is(err, serverAppdbAnalysis.ErrNotFound) {
		return analysisRequest, ErrNoPendingRequest
	}
	return analysisRequest, err
}
