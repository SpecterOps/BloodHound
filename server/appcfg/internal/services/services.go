// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/server/audit"
)

var (
	ErrNotFound = errors.New("not found")
)

type Database interface {
	GetDatapipeStatus(ctx context.Context) (DatapipeStatus, error)
	GetConfigurationParameter(ctx context.Context, parameterKey ParameterKey) (Parameter, error)
	GetAllConfigurationParameters(ctx context.Context) (Parameters, error)
	UpsertConfigurationParameter(ctx context.Context, parameterKey ParameterKey, value types.JSONBObject) error
}

type Auditor interface {
	Begin(ctx context.Context, entry audit.Entry) (uuid.UUID, error)
	Complete(ctx context.Context, commitID uuid.UUID, entry audit.Entry, operationErr error) error
}

// Reader provides application configuration read operations without requiring
// dependencies used only by mutating operations.
type Reader struct {
	db Database
}

type Service struct {
	*Reader
	auditor Auditor
}

func NewReader(databaseInterface Database) *Reader {
	if databaseInterface == nil {
		panic("appcfg: reader requires a non-nil Database")
	}

	return &Reader{db: databaseInterface}
}

func NewService(databaseInterface Database, auditor Auditor) *Service {
	if databaseInterface == nil {
		panic("appcfg: service requires a non-nil Database")
	}
	if auditor == nil {
		panic("appcfg: service requires a non-nil Auditor")
	}

	return &Service{Reader: &Reader{db: databaseInterface}, auditor: auditor}
}
