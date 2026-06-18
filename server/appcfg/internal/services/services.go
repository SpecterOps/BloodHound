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

import (
	"context"
	"errors"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
)

//go:generate go tool mockery

// DatapipeStatusType represents the current status of the datapipe.
type DatapipeStatusType string

const (
	DatapipeStatusIdle      DatapipeStatusType = "idle"
	DatapipeStatusIngesting DatapipeStatusType = "ingesting"
	DatapipeStatusAnalyzing DatapipeStatusType = "analyzing"
	DatapipeStatusPurging   DatapipeStatusType = "purging"
	DatapipeStatusPruning   DatapipeStatusType = "pruning"
	DatapipeStatusStarting  DatapipeStatusType = "starting"
)

// DatapipeStatus represents the current state of the datapipe.
type DatapipeStatus struct {
	Status                  DatapipeStatusType
	UpdatedAt               time.Time
	LastCompleteAnalysisAt  null.Time
	LastAnalysisRunAt       null.Time
	NextScheduledAnalysisAt null.Time
}

var (
	ErrNotFound = errors.New("not found")
)

type Database interface {
	GetDatapipeStatus(ctx context.Context) (DatapipeStatus, error)
}

type Service struct {
	db Database
}

func NewService(db Database) *Service {
	return &Service{db: db}
}

// GetDatapipeService returns the current datapipe status
func (s *Service) GetDatapipeStatus(ctx context.Context) (DatapipeStatus, error) {
	return s.db.GetDatapipeStatus(ctx)
}
