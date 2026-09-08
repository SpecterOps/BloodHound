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
)

var (
	ErrNotFound = errors.New("not found")
)

type Database interface {
	GetDatapipeStatus(ctx context.Context) (DatapipeStatus, error)
	GetAllConfigurationParameters(ctx context.Context) (Parameters, error)
	GetConfigurationParameter(ctx context.Context, parameterKey ParameterKey) (Parameter, error)
	GetAllConfigurationParameters(ctx context.Context) (Parameters, error)
}

type Service struct {
	db Database
}

func NewService(databaseInterface Database) *Service {
	return &Service{db: databaseInterface}
}
