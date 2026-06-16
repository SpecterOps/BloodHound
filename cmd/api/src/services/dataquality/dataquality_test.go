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

package dataquality

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dataquality/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOpenGraphDataQualitySourceKinds(t *testing.T) {
	var (
		ctx                = context.Background()
		firstEnvID   int32 = 11
		secondEnvID  int32 = 12
		firstKindID  int32 = 21
		secondKindID int32 = 22
	)

	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDataQualityData(ctrl)

	environments := []model.SchemaEnvironment{
		{Serial: model.Serial{ID: firstEnvID}, SourceKindId: firstKindID},
		{Serial: model.Serial{ID: secondEnvID}, SourceKindId: secondKindID},
	}
	sourceKinds := []model.Kind{
		{ID: firstKindID, Name: "Domain"},
		{ID: secondKindID, Name: "Tenant"},
	}

	mockDB.EXPECT().GetKindsByIDs(ctx, firstKindID, secondKindID).Return(sourceKinds, nil)

	sourceKindsByEnvironmentID, err := openGraphDataQualitySourceKinds(ctx, mockDB, environments)

	require.NoError(t, err)
	require.Equal(t, model.Kind{ID: firstKindID, Name: "Domain"}, sourceKindsByEnvironmentID[firstEnvID])
	require.Equal(t, model.Kind{ID: secondKindID, Name: "Tenant"}, sourceKindsByEnvironmentID[secondEnvID])
}

func TestOpenGraphDataQualitySourceKinds_MissingKind(t *testing.T) {
	var (
		ctx              = context.Background()
		envID      int32 = 11
		sourceKind int32 = 21
	)

	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDataQualityData(ctrl)

	environments := []model.SchemaEnvironment{
		{Serial: model.Serial{ID: envID}, SourceKindId: sourceKind},
	}

	mockDB.EXPECT().GetKindsByIDs(ctx, sourceKind).Return([]model.Kind{}, nil)

	_, err := openGraphDataQualitySourceKinds(ctx, mockDB, environments)

	require.ErrorContains(t, err, "source kind 21 not found for schema environment 11")
}
