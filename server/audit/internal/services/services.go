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
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type Entry struct {
	Action model.AuditLogAction
	Fields map[string]any
}

type Log struct {
	CreatedAt       time.Time
	ActorID         string
	ActorName       string
	ActorEmail      string
	Action          model.AuditLogAction
	Fields          map[string]any
	RequestID       string
	SourceIPAddress string
	Status          model.AuditLogEntryStatus
	CommitID        uuid.UUID
}

type Database interface {
	CreateAuditLog(ctx context.Context, auditLog Log) error
}

type Service struct {
	database         Database
	identityResolver auth.IdentityResolver
}

func NewService(databaseInterface Database, identityResolver auth.IdentityResolver) *Service {
	if databaseInterface == nil {
		panic("audit: service requires a non-nil Database")
	}
	if identityResolver == nil {
		panic("audit: service requires a non-nil IdentityResolver")
	}

	return &Service{database: databaseInterface, identityResolver: identityResolver}
}

func (s *Service) Begin(ctx context.Context, entry Entry) (uuid.UUID, error) {
	var (
		commitID, err = uuid.NewV4()
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("generating audit commit id: %w", err)
	}

	if err = s.database.CreateAuditLog(ctx, s.newLog(ctx, commitID, entry, model.AuditLogStatusIntent, nil)); err != nil {
		return uuid.Nil, fmt.Errorf("creating audit intent: %w", err)
	}

	return commitID, nil
}

func (s *Service) Complete(ctx context.Context, commitID uuid.UUID, entry Entry, operationErr error) error {
	var status = model.AuditLogStatusSuccess
	if operationErr != nil {
		status = model.AuditLogStatusFailure
	}

	if err := s.database.CreateAuditLog(ctx, s.newLog(ctx, commitID, entry, status, operationErr)); err != nil {
		return fmt.Errorf("creating audit %s outcome: %w", status, err)
	}

	return nil
}

func (s *Service) newLog(ctx context.Context, commitID uuid.UUID, entry Entry, status model.AuditLogEntryStatus, operationErr error) Log {
	var (
		requestContext = bhctx.Get(ctx)
		fields         = make(map[string]any, len(entry.Fields)+1)
		auditLog       = Log{
			CreatedAt:       time.Now().UTC(),
			Action:          entry.Action,
			Fields:          fields,
			RequestID:       requestContext.RequestID,
			SourceIPAddress: requestContext.RequestIP,
			Status:          status,
			CommitID:        commitID,
		}
	)

	for key, value := range entry.Fields {
		fields[key] = value
	}
	if operationErr != nil {
		fields["error"] = operationErr.Error()
	}

	if identity, err := s.identityResolver.GetIdentity(requestContext.AuthCtx); err == nil {
		auditLog.ActorID = identity.ID.String()
		auditLog.ActorName = identity.Name
		auditLog.ActorEmail = identity.Email
	}

	return auditLog
}
