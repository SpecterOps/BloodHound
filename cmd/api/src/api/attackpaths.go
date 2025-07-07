// Copyright 2025 Specter Ops, Inc.
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

package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func ParseAssetGroupTagIdWithFallback(ctx context.Context, db database.Database, maybeAssetGroupTagId string) (int, error) {
	if maybeAssetGroupTagId != "" {
		if tierIdParam, err := strconv.Atoi(maybeAssetGroupTagId); err != nil {
			return 0, err
		} else if tierIdParam == 0 {
			// This is a workaround to supply tiering agnostic findings
			return 0, nil
		} else if _, err = db.GetAssetGroupTag(ctx, tierIdParam); err != nil {
			return 0, err
		} else {
			return tierIdParam, nil
		}
	}
	// Fallback to tier zero if not supplied
	if agt, err := db.GetAssetGroupTags(ctx, model.SQLFilter{SQLString: "type = ? AND position = ?", Params: []any{model.AssetGroupTagTypeTier, model.AssetGroupTierZeroPosition}}); err != nil {
		return 0, err
	} else if len(agt) == 0 {
		return 0, fmt.Errorf("no asset group tag found for tier zero")
	} else {
		return agt[0].ID, nil
	}
}
