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

//go:build integration

package database_test

import (
	"context"
	"slices"
	"testing"

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/bloodhound/src/utils/test"
)

func TestCreateGetUpdateDeleteAssetGroup(t *testing.T) {
	var (
		dbInst        = integration.SetupDB(t)
		testCtx       = context.Background()
		newAssetGroup model.AssetGroup
		err           error
	)

	if newAssetGroup, err = dbInst.CreateAssetGroup(testCtx, "test asset group", "test", false); err != nil {
		t.Fatalf("Error creating asset group: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionCreateAssetGroup, "asset_group_name", newAssetGroup.Name); err != nil {
		t.Fatalf("Error verifying CreateAssetGroup audit logs:\n%v", err)
	}

	if assetGroups, err := dbInst.GetAllAssetGroups(testCtx, "", model.SQLFilter{}); err != nil {
		t.Fatalf("Error retrieving asset groups: %v", err)
	} else if !slices.ContainsFunc(assetGroups, func(ag model.AssetGroup) bool { return ag.Name == "test asset group" }) {
		t.Fatalf("Created asset group not returned:\n%#v", assetGroups)
	}

	updatedAssetGroup := model.AssetGroup{
		Serial: model.Serial{
			ID: newAssetGroup.ID,
		},
		Name:        "updated asset group",
		Tag:         newAssetGroup.Tag,
		SystemGroup: newAssetGroup.SystemGroup,
		Selectors:   newAssetGroup.Selectors,
		Collections: newAssetGroup.Collections,
		MemberCount: newAssetGroup.MemberCount,
	}
	if err = dbInst.UpdateAssetGroup(testCtx, updatedAssetGroup); err != nil {
		t.Fatalf("Error updating asset group: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionUpdateAssetGroup, "asset_group_name", "updated asset group"); err != nil {
		t.Fatalf("Error veriying UpdateAssetGroup audit logs:\n%v", err)
	}

	if err = dbInst.DeleteAssetGroup(testCtx, updatedAssetGroup); err != nil {
		t.Fatalf("Error deleting asset group: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionDeleteAssetGroup, "asset_group_name", "updated asset group"); err != nil {
		t.Fatalf("Error veriying DeleteAssetGroup audit logs:\n%v", err)
	}
}
