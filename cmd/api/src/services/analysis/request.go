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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/analysis.go -package=mocks . Service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
)

// Service is the single entry point for queueing, fetching, and clearing
// analysis pipeline requests. API handlers and background jobs that need to
// trigger analysis should go through this service rather than calling the
// database layer directly; the service owns the precedence rules that decide
// what happens when multiple requests collide.
type Service interface {
	GetAnalysisRequest(ctx context.Context) (model.AnalysisRequest, error)
	DeleteAnalysisRequest(ctx context.Context) error
	RequestAnalysisForAGT(ctx context.Context, requestedBy string) error
	RequestFullAnalysis(ctx context.Context, requestedBy string) error
	RequestGraphDataDeletion(ctx context.Context, request model.AnalysisRequest) error
}

type service struct {
	db database.Database
}

// NewAnalysisService returns a Service backed by the provided database.
func NewAnalysisService(db database.Database) Service {
	return &service{db: db}
}

func (s *service) GetAnalysisRequest(ctx context.Context) (model.AnalysisRequest, error) {
	return s.db.GetAnalysisRequest(ctx)
}

func (s *service) DeleteAnalysisRequest(ctx context.Context) error {
	return s.db.DeleteAnalysisRequest(ctx)
}

// RequestAnalysisForAGT queues an analysis on behalf of an asset group tag
// mutation. When FeatureAGTPartialAnalysis is enabled the request runs the
// partial pipeline (tagging through findings); otherwise it runs the full
// pipeline.
func (s *service) RequestAnalysisForAGT(ctx context.Context, requestedBy string) error {
	var step = model.AnalysisStepAll
	if appcfg.GetAGTPartialAnalysisEnabled(ctx, s.db) {
		step = model.AnalysisStepTaggingToCompletion
	}
	return s.applyRequest(ctx, buildAnalysisRequest(requestedBy, step))
}

// RequestFullAnalysis queues a full-pipeline analysis.
// This is used for user-initiated analysis and for init at startup.
func (s *service) RequestFullAnalysis(ctx context.Context, requestedBy string) error {
	return s.applyRequest(ctx, buildAnalysisRequest(requestedBy, model.AnalysisStepAll))
}

// RequestGraphDataDeletion queues a collected graph data deletion. Deletion is
// sticky and overrides any pending analysis request.
func (s *service) RequestGraphDataDeletion(ctx context.Context, request model.AnalysisRequest) error {
	request.RequestType = model.AnalysisRequestDeletion
	return s.applyRequest(ctx, request)
}

// buildAnalysisRequest constructs an AnalysisRequest struct for the analysis
// pipeline. It does not touch the database; applyRequest is what persists it.
func buildAnalysisRequest(requestedBy string, step model.AnalysisStep) model.AnalysisRequest {
	return model.AnalysisRequest{
		RequestType:  model.AnalysisRequestAnalysis,
		RequestedBy:  requestedBy,
		AnalysisStep: step,
	}
}

// applyRequest reads any existing request, reconciles it with the incoming one
// via the pure merge helper, and writes the resulting decision. To keep the
// merge rules in exactly one place this bypasses the database layer's
// precedence branch by deleting the existing row before writing the merged
// request.
func (s *service) applyRequest(ctx context.Context, incoming model.AnalysisRequest) error {
	var (
		existing    model.AnalysisRequest
		hasExisting = true
	)
	if got, err := s.db.GetAnalysisRequest(ctx); err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, database.ErrNotFound) {
			return err
		}
		hasExisting = false
	} else {
		existing = got
	}

	decision := merge(existing, hasExisting, incoming)
	if !decision.write {
		return nil
	}

	if hasExisting {
		if err := s.db.DeleteAnalysisRequest(ctx); err != nil {
			return err
		}
	}

	if decision.merged.RequestType == model.AnalysisRequestDeletion {
		return s.db.RequestCollectedGraphDataDeletion(ctx, decision.merged)
	}
	return s.db.RequestAnalysis(ctx, decision.merged.RequestedBy, decision.merged.AnalysisStep)
}
