// Copyright 2024 Specter Ops, Inc.
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

package database

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"gorm.io/gorm"
)

type AnalysisRequestData interface {
	DeleteAnalysisRequest(ctx context.Context) error
	GetAnalysisRequest(ctx context.Context) (model.AnalysisRequest, error)
	HasAnalysisRequest(ctx context.Context) bool
	HasCollectedGraphDataDeletionRequest(ctx context.Context) (model.AnalysisRequest, bool)
	RequestAnalysis(ctx context.Context, requester string, analysisMode model.AnalysisMode) error
	RequestCollectedGraphDataDeletion(ctx context.Context, request model.AnalysisRequest) error
}

func (s *BloodhoundDB) DeleteAnalysisRequest(ctx context.Context) error {
	tx := s.db.WithContext(ctx).Exec(`truncate analysis_request_switch;`)
	return tx.Error
}

func (s *BloodhoundDB) GetAnalysisRequest(ctx context.Context) (model.AnalysisRequest, error) {
	var analysisRequest model.AnalysisRequest

	tx := s.db.WithContext(ctx).Select("requested_by, request_type, requested_at, analysis_step").Table("analysis_request_switch").First(&analysisRequest)

	return analysisRequest, CheckError(tx)
}

func (s *BloodhoundDB) HasAnalysisRequest(ctx context.Context) bool {
	var exists bool

	tx := s.db.WithContext(ctx).Raw(`select exists(select * from analysis_request_switch where request_type = ? limit 1);`, model.AnalysisRequestAnalysis).Scan(&exists)
	if tx.Error != nil {
		slog.ErrorContext(ctx, "Error determining if there's an analysis request", attr.Error(tx.Error))
	}
	return exists
}

func (s *BloodhoundDB) HasCollectedGraphDataDeletionRequest(ctx context.Context) (model.AnalysisRequest, bool) {
	var record model.AnalysisRequest

	tx := s.db.WithContext(ctx).Raw(`select * from analysis_request_switch where request_type = ? limit 1;`, model.AnalysisRequestDeletion).First(&record)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return record, false
		}
		slog.ErrorContext(ctx, "Error querying deletion request", attr.Error(tx.Error))
		return record, false
	}
	return record, true
}

// setAnalysisRequest inserts a row into analysis_request_switch for both a collected graph data deletion request or an analysis request.
// There should only ever be 1 row.
// If an analysis request is already queued and another analysis request comes in, the requested analysis steps are merged
// (bitwise OR) so subsequent requests can only widen the work, never narrow it.
// If an analysis request is present when a deletion request comes in, that overwrites the analysis to deletion but not vice-versa.
// To request: Use the helper methods `RequestAnalysis` and `RequestCollectedGraphDataDeletion`
func (s *BloodhoundDB) setAnalysisRequest(ctx context.Context, request model.AnalysisRequest) error {
	var (
		now                 = time.Now().UTC()
		analysisRequestType = model.AnalysisRequestAnalysis
		args                = []any{
			analysisRequestType,
			request.RequestedBy,
			request.RequestType,
			now,
			request.AnalysisSteps,
			request.DeleteAllGraph,
			request.DeleteSourcelessGraph,
			pq.StringArray(request.DeleteSourceKinds),
			pq.StringArray(request.DeleteRelationships),
		}

		// This upsert keeps the singleton request row atomic under concurrent requests.
		// The CTE passes the analysis request type once so the SQL does not hard-code the model value.
		// On conflict, only existing analysis requests may be updated. Incoming analysis requests merge
		// step bits while preserving original audit/deletion fields, incoming deletion requests overwrite
		// analysis requests, and existing deletion requests remain unchanged.
		upsertSQL = `
			WITH request_constants AS (
				SELECT ?::text AS analysis_request_type
			)
			INSERT INTO analysis_request_switch (
					requested_by,
					request_type,
					requested_at,
					analysis_step,
					delete_all_graph,
					delete_sourceless_graph,
					delete_source_kinds,
					delete_relationships
				)
			VALUES (?, ?, ?, ?, ?, ?, ?::text[], ?::text[])
			ON CONFLICT (singleton) DO UPDATE
			SET
				requested_by = CASE
					WHEN EXCLUDED.request_type = analysis_request_switch.request_type THEN analysis_request_switch.requested_by
					ELSE EXCLUDED.requested_by
				END,
				request_type = EXCLUDED.request_type,
				requested_at = CASE
					WHEN EXCLUDED.request_type = analysis_request_switch.request_type THEN analysis_request_switch.requested_at
					ELSE EXCLUDED.requested_at
				END,
				analysis_step = CASE
					WHEN EXCLUDED.request_type = analysis_request_switch.request_type THEN COALESCE(analysis_request_switch.analysis_step, 0) | COALESCE(EXCLUDED.analysis_step, 0)
					ELSE EXCLUDED.analysis_step
				END,
				delete_all_graph = CASE
					WHEN EXCLUDED.request_type = analysis_request_switch.request_type THEN analysis_request_switch.delete_all_graph
					ELSE EXCLUDED.delete_all_graph
				END,
				delete_sourceless_graph = CASE
					WHEN EXCLUDED.request_type = analysis_request_switch.request_type THEN analysis_request_switch.delete_sourceless_graph
					ELSE EXCLUDED.delete_sourceless_graph
				END,
				delete_source_kinds = CASE
					WHEN EXCLUDED.request_type = analysis_request_switch.request_type THEN analysis_request_switch.delete_source_kinds
					ELSE EXCLUDED.delete_source_kinds
				END,
				delete_relationships = CASE
					WHEN EXCLUDED.request_type = analysis_request_switch.request_type THEN analysis_request_switch.delete_relationships
					ELSE EXCLUDED.delete_relationships
				END
			WHERE analysis_request_switch.request_type = (SELECT analysis_request_type FROM request_constants)
				AND (
					EXCLUDED.request_type <> (SELECT analysis_request_type FROM request_constants)
					OR COALESCE(analysis_request_switch.analysis_step, 0) <> (COALESCE(analysis_request_switch.analysis_step, 0) | COALESCE(EXCLUDED.analysis_step, 0))
				);`
	)

	return s.db.WithContext(ctx).Exec(upsertSQL, args...).Error
}

// RequestAnalysis will request an analysis be executed, as long as there isn't an existing analysis request or collected graph data deletion request, then it no-ops
func (s *BloodhoundDB) RequestAnalysis(ctx context.Context, requestedBy string, analysisMode model.AnalysisMode) error {
	var steps = analysisMode.AnalysisStepsFromMode()

	if !appcfg.GetVariableAnalysisModeEnabled(ctx, s) {
		steps = model.AnalysisStepsFull()
	}

	return s.setAnalysisRequest(ctx, model.AnalysisRequest{RequestType: model.AnalysisRequestAnalysis, RequestedBy: requestedBy, AnalysisSteps: steps})
}

// RequestCollectedGraphDataDeletion will request collected graph data be deleted, if an analysis request is present, it will overwrite that.
func (s *BloodhoundDB) RequestCollectedGraphDataDeletion(ctx context.Context, request model.AnalysisRequest) error {
	slog.InfoContext(ctx, "Request collected graph data deletion", slog.String("requested_by", request.RequestedBy))
	return s.setAnalysisRequest(ctx, request)
}
