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
package endpoint_test

import (
	"context"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	graph_mocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResolver_StartDone(t *testing.T) {
	var (
		ctx, done = context.WithTimeout(context.Background(), time.Second*10)
		mockCtrl  = gomock.NewController(t)
		mockGraph = graph_mocks.NewMockDatabase(mockCtrl)
		resolver  = endpoint.NewResolver(mockGraph)
	)

	defer done()

	// Mock out the read transaction calls
	mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil).AnyTimes()

	// Ensure that start and done are reentrant-safe
	resolver.Start(ctx, 4)
	resolver.Start(ctx, 4)

	assert.NoError(t, resolver.Done())
	assert.NoError(t, resolver.Done())
}
