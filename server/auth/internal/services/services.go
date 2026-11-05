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
)

type Database interface {
	Logout(ctx context.Context, sessionID int64) error
}

// Service implements the auth use cases on top of a Database implementation.
type Service struct {
	db Database
}

// NewService constructs a Service backed by the supplied Database implementation.
func NewService(databaseInterface Database) *Service {
	if databaseInterface == nil {
		panic("auth: service requires a non-nil Database")
	}

	return &Service{db: databaseInterface}
}

func (s *Service) Logout(ctx context.Context, sessionID int64) error {
	return s.db.Logout(ctx, sessionID)
}
