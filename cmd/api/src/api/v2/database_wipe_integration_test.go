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

package v2_test

import (
	"context"
	"testing"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/stretchr/testify/require"
)

func TestBuildDeleteRequest(t *testing.T) {
	var (
		testCtx   = context.Background()
		graphDB   = integration.OpenGraphDB(t, graphschema.DefaultGraphSchema())
		db        = integration.SetupDB(t)
		resources = v2.Resources{
			DB:    db,
			Graph: graphDB,
		}
	)
	defer graphDB.Close(testCtx)

	// successfully maps to kind names
	request, err := resources.BuildDeleteRequest(testCtx, "1234", v2.DatabaseWipe{DeleteSourceKinds: []int{1, 2}})
	require.Nil(t, err)
	require.Len(t, request.DeleteSourceKinds, 2)

	// successfully maps sourceless
	request, err = resources.BuildDeleteRequest(testCtx, "1234", v2.DatabaseWipe{DeleteSourceKinds: []int{0, 1, 2}})
	require.Nil(t, err)
	require.True(t, request.DeleteSourcelessGraph)
	require.Len(t, request.DeleteSourceKinds, 2)

	// payload contains a sourceKindID that doesn't exist
	request, err = resources.BuildDeleteRequest(testCtx, "1234", v2.DatabaseWipe{DeleteSourceKinds: []int{1, 2, 5}})
	require.Error(t, err)
	require.ErrorContains(t, err, "requested source kind id 5 not found")
}
