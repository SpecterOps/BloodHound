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
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
)

var ErrAuditLog = errors.New("audit log operation failed")

const auditLogOutcomeTimeout = 30 * time.Second

type Action string

const (
	ActionUpdateParameter              Action = "UpdateParameter"
	ActionToggleEarlyAccessFeatureFlag Action = "ToggleEarlyAccessFeatureFlag"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusIntent  Status = "intent"
)

type AuditLogEntry struct {
	Action Action
	Fields map[string]any
}

type AuditLog struct {
	ActorID         string
	ActorName       string
	ActorEmail      string
	Action          Action
	Fields          map[string]any
	RequestID       string
	SourceIPAddress string
	Status          Status
	CommitID        uuid.UUID
}

type Database interface {
	CreateAuditLog(ctx context.Context, auditLog AuditLog) error
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

func (s *Service) Begin(ctx context.Context, entry AuditLogEntry) (uuid.UUID, error) {
	var (
		commitID, err = uuid.NewV4()
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: generating audit commit id: %s", ErrAuditLog, err)
	}

	if err = s.database.CreateAuditLog(ctx, s.newLog(ctx, commitID, entry, StatusIntent, nil)); err != nil {
		return uuid.Nil, fmt.Errorf("%w: creating audit intent: %s", ErrAuditLog, err)
	}

	return commitID, nil
}

func (s *Service) Complete(ctx context.Context, commitID uuid.UUID, entry AuditLogEntry, operationErr error) error {
	var (
		status               = StatusSuccess
		auditContext, cancel = context.WithTimeout(context.WithoutCancel(ctx), auditLogOutcomeTimeout)
	)
	defer cancel()

	if operationErr != nil {
		status = StatusFailure
	}

	if err := s.database.CreateAuditLog(auditContext, s.newLog(auditContext, commitID, entry, status, operationErr)); err != nil {
		return fmt.Errorf("%w: creating audit %s outcome: %s", ErrAuditLog, status, err)
	}

	return nil
}

func (s *Service) newLog(ctx context.Context, commitID uuid.UUID, entry AuditLogEntry, status Status, operationErr error) AuditLog {
	var (
		requestContext = bhctx.Get(ctx)
		fields         = make(map[string]any, len(entry.Fields)+1)
		auditLog       = AuditLog{
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
