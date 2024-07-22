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
	"database/sql"
	"time"

	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
)

type AnalysisRequestData interface {
	DeleteAnalysisRequest(ctx context.Context) error
	GetAnalysisRequest(ctx context.Context) (model.AnalysisRequest, error)
	HasAnalysisRequest(ctx context.Context) bool
	HasCollectedGraphDataDeletionRequest(ctx context.Context) bool
	RequestAnalysis(ctx context.Context, requester string) error
	RequestCollectedGraphDataDeletion(ctx context.Context, requester string) error
}

func (s *BloodhoundDB) DeleteAnalysisRequest(ctx context.Context) error {
	tx := s.db.WithContext(ctx).Exec(`truncate analysis_request_switch;`)
	return tx.Error
}

func (s *BloodhoundDB) GetAnalysisRequest(ctx context.Context) (model.AnalysisRequest, error) {
	var analysisRequest model.AnalysisRequest

	// Note: GORM Raw does not throw any errors if no row is found. We can inspect rows affected as a workaround
	if tx := s.db.WithContext(ctx).Raw(`select requested_by, request_type, requested_at from analysis_request_switch limit 1;`).Scan(&analysisRequest); tx.RowsAffected == 0 {
		return analysisRequest, sql.ErrNoRows
	}

	return analysisRequest, nil
}

func (s *BloodhoundDB) HasAnalysisRequest(ctx context.Context) bool {
	var exists bool

	tx := s.db.WithContext(ctx).Raw(`select exists(select * from analysis_request_switch where request_type = ? limit 1);`, model.AnalysisRequestAnalysis).Scan(&exists)
	if tx.Error != nil {
		log.Errorf("Error determining if there's an analysis request: %v", tx.Error)
	}
	return exists
}

func (s *BloodhoundDB) HasCollectedGraphDataDeletionRequest(ctx context.Context) bool {
	var exists bool

	tx := s.db.WithContext(ctx).Raw(`select exists(select * from analysis_request_switch where request_type = ? limit 1);`, model.AnalysisRequestDeletion).Scan(&exists)
	if tx.Error != nil {
		log.Errorf("Error determining if there's a deletion request: %v", tx.Error)
	}
	return exists
}

// This inserts a row into analysis_request_switch for both a collected graph data deletion request or an analysis request.
// There should only ever be 1 row, if a request is present, subsequent requests no-op
// If an analysis request is present when a deletion request comes in, that overwrites the analysis to deletion but not vice-versa
// To request: Use the helper methods `RequestAnalysis` and `RequestCollectedGraphDataDeletion`
func (s *BloodhoundDB) setAnalysisRequest(ctx context.Context, requestType model.AnalysisRequestType, requestedBy string) error {
	if analReq, err := s.GetAnalysisRequest(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	} else if errors.Is(err, sql.ErrNoRows) {
		// Analysis request doesn't exist so insert one
		insertSql := `insert into analysis_request_switch (requested_by, request_type, requested_at) values (?, ?, ?);`
		tx := s.db.WithContext(ctx).Exec(insertSql, requestedBy, requestType, time.Now().UTC())
		return tx.Error
	} else {
		// Analysis request existed, we only want to overwrite if request is for a deletion request, otherwise ignore additional requests
		if analReq.RequestType == model.AnalysisRequestAnalysis && requestType == model.AnalysisRequestDeletion {
			updateSql := `update analysis_request_switch set requested_by = ?, request_type = ?, requested_at = ? limit 1;`
			tx := s.db.WithContext(ctx).Exec(updateSql, requestedBy, requestType, time.Now().UTC())
			return tx.Error
		}
		return nil
	}
}

// This will request an analysis be executed, as long as there isn't an existing analysis request or collected graph data deletion request, then it no-ops
func (s *BloodhoundDB) RequestAnalysis(ctx context.Context, requestedBy string) error {
	log.Infof("Analysis requested by %s", requestedBy)
	return s.setAnalysisRequest(ctx, model.AnalysisRequestAnalysis, requestedBy)
}

// This will request collected graph data be deleted, if an analysis request is present, it will overwrite that.
func (s *BloodhoundDB) RequestCollectedGraphDataDeletion(ctx context.Context, requestedBy string) error {
	log.Infof("Collected graph data deletion requested by %s", requestedBy)
	return s.setAnalysisRequest(ctx, model.AnalysisRequestDeletion, requestedBy)
}
