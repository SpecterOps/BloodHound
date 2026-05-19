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

package services

//go:generate go tool mockery

import (
	"context"
	"errors"
	"time"
)

// RequestedAnalysisType identifies the category of work an analysis request represents.
type RequestedAnalysisType string

const (
	RequestedAnalysisTypeAnalysis RequestedAnalysisType = "analysis"
	RequestedAnalysisTypeDeletion RequestedAnalysisType = "deletion"
)

// RequestedAnalysis is the domain representation of a pending analysis request.
type RequestedAnalysis struct {
	RequestedBy string
	RequestType RequestedAnalysisType
	RequestedAt time.Time
	// Deletes all nodes and edges in the graph
	DeleteAllGraph bool
	// Deletes all nodes and edges in the graph that have a type not registered in the source_kinds table
	DeleteSourcelessGraph bool
	// Deletes all nodes and edges per kind provided.
	DeleteSourceKinds []string
	// Deletes all relationships by name
	DeleteRelationships []string
}

// ErrNotFound is the sentinel that Database implementations must return when no
// analysis request row exists. Defining it here (on the consumer side) keeps the
// appdb store free of any import back into this package.
var ErrNotFound = errors.New("analysis request not found")

// ErrNoPendingRequest indicates that there is no analysis request currently pending.
var ErrNoPendingRequest = errors.New("no pending analysis request")

// Database describes the persistence capabilities the analysis Service requires. Implementations
// are expected to translate driver- or ORM-specific not-found errors into appdb-level sentinels
// so that the Service can map them to its own failure-mode errors.
type Database interface {
	GetAnalysisRequest(ctx context.Context) (RequestedAnalysis, error)
	CreateAnalysisRequest(ctx context.Context, requestedBy string) (RequestedAnalysis, bool, error)
}

// Service implements the analysis use cases on top of a Database implementation.
type Service struct {
	db Database
}

// NewService constructs a Service backed by the supplied Database implementation.
func NewService(databaseInterface Database) *Service {
	return &Service{db: databaseInterface}
}

// GetRequest returns the currently pending analysis request. ErrNoPendingRequest is returned
// when no request is pending; any other error indicates a failure servicing the request.
func (s *Service) GetRequest(ctx context.Context) (RequestedAnalysis, error) {
	analysisRequest, err := s.db.GetAnalysisRequest(ctx)
	if errors.Is(err, ErrNotFound) {
		return analysisRequest, ErrNoPendingRequest
	}
	return analysisRequest, err
}

// CreateRequest submits a new analysis request attributed to the given user. The currently
// pending request is returned along with a boolean indicating whether this call created it
// (true) or a request was already pending (false).
func (s *Service) CreateRequest(ctx context.Context, requestedBy string) (RequestedAnalysis, bool, error) {
	return s.db.CreateAnalysisRequest(ctx, requestedBy)
}
