// Copyright 2023 Specter Ops, Inc.
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

//go:build integration
// +build integration

package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
)

func TestDatabase_ListAuditLogs(t *testing.T) {
	var (
		dbInst = integration.OpenDatabase(t)

		auditLogIdFilter = model.QueryParameterFilter{
			Name:         "id",
			Operator:     model.GreaterThan,
			Value:        "4",
			IsStringData: false,
		}
		auditLogIdFilterMap = model.QueryParameterFilterMap{auditLogIdFilter.Name: model.QueryParameterFilters{auditLogIdFilter}}
	)

	if err := integration.Prepare(dbInst); err != nil {
		t.Fatalf("Failed preparing DB: %v", err)
	}

	mockCtx := ctx.Context{
		RequestID: "requestID",
		AuthCtx: auth.Context{
			Owner:   model.User{},
			Session: model.UserSession{},
		},
	}
	for i := 0; i < 7; i++ {
		if err := dbInst.AppendAuditLog(ctx.Set(context.Background(), &mockCtx), model.AuditEntry{Model: &model.User{}, Action: "CreateUser", Status: model.AuditStatusSuccess}); err != nil {
			t.Fatalf("Error creating audit log: %v", err)
		}
	}

	if _, count, err := dbInst.ListAuditLogs(time.Now(), time.Now(), 0, 10, "", model.SQLFilter{}); err != nil {
		t.Fatalf("Failed to list all audit logs: %v", err)
	} else if count != 7 {
		t.Fatalf("Expected 7 audit logs to be returned")
	} else if filter, err := auditLogIdFilterMap.BuildSQLFilter(); err != nil {
		t.Fatalf("Failed to generate SQL Filter: %v", err)
		// Limit is set to 1 to verify that count is total filtered count, not response size
	} else if _, count, err = dbInst.ListAuditLogs(time.Now(), time.Now(), 0, 1, "", filter); err != nil {
		t.Fatalf("Failed to list filtered events: %v", err)
	} else if count != 3 {
		t.Fatalf("Expected 3 audit logs to be returned")
	}
}
