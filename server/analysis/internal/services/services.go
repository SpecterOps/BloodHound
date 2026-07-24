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

	"github.com/specterops/bloodhound/server/alerts"
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

// ErrNoPendingRequest indicates that there is no analysis request currently pending.
var ErrNoPendingRequest = errors.New("no pending analysis request")

// ErrDeletionRequestPending is returned when a cancel is attempted against a deletion request.
// A deletion request cannot be cancelled; only an analysis request may be withdrawn.
var ErrDeletionRequestPending = errors.New("a deletion request is pending and cannot be cancelled")

// Database describes the persistence capabilities the analysis Service requires. Implementations
// are expected to translate driver- or ORM-specific not-found errors into appdb-level sentinels
// so that the Service can map them to its own failure-mode errors.
type Database interface {
	GetAnalysisRequest(ctx context.Context) (RequestedAnalysis, error)
	CreateAnalysisRequest(ctx context.Context, requestedBy string) (RequestedAnalysis, bool, error)
	DeleteAnalysisRequest(ctx context.Context) error
}

// Service implements the analysis use cases on top of a Database implementation & event publisher.
type Service struct {
	db        Database
	publisher alerts.Publisher
}

// NewService constructs a Service backed by the supplied Database implementation & event publisher.
func NewService(databaseInterface Database, eventPublisher alerts.Publisher) *Service {
	if eventPublisher == nil {
		return &Service{db: databaseInterface, publisher: alerts.NewNoopPubSub()}
	}
	return &Service{db: databaseInterface, publisher: eventPublisher}
}

// GetRequest returns the currently pending analysis request. ErrNoPendingRequest is returned
// when no request is pending; any other error indicates a failure servicing the request.
func (s *Service) GetRequest(ctx context.Context) (RequestedAnalysis, error) {
	return s.db.GetAnalysisRequest(ctx)
}

// CreateRequest submits a new analysis request attributed to the given user. The currently
// pending request is returned along with a boolean indicating whether this call created it
// (true) or a request was already pending (false).
func (s *Service) CreateRequest(ctx context.Context, requestedBy string) (RequestedAnalysis, bool, error) {
	if analysisRequest, success, err := s.db.CreateAnalysisRequest(ctx, requestedBy); err != nil {
		s.publisher.Publish(ctx, "analysis.request.error", alerts.CreateAlertEventInput{Message: "Error Requesting Analysis", Data: map[string]any{"error": err}})
		return analysisRequest, success, err
	} else if !success {
		s.publisher.Publish(ctx, "analysis.request.failure", alerts.CreateAlertEventInput{Message: "Failed to Request Analysis; Analysis Already Requested", Data: map[string]any{"error": err}})
		return analysisRequest, success, err
	} else {
		s.publisher.Publish(ctx, "analysis.request.success", alerts.CreateAlertEventInput{Message: "Requesting Analysis Successful", Data: map[string]any{"requested_by": requestedBy}})
		return analysisRequest, success, err

	}
}

func (s *Service) CancelAnalysisRequest(ctx context.Context) error {
	return s.db.DeleteAnalysisRequest(ctx)
}
