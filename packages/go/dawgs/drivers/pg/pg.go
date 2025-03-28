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

package pg

import (
	"context"
	"time"

	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/graph"
)

const (
	DriverName = "pg"

	defaultTransactionTimeout = time.Minute * 15
	// defaultBatchWriteSize is currently set to 2k. This is meant to strike a balance between the cost of thousands
	// of round-trips against the cost of locking tables for too long.
	defaultBatchWriteSize = 2_000
)

func init() {
	dawgs.Register(DriverName, func(ctx context.Context, cfg dawgs.Config) (graph.Database, error) {
		return NewDriver(cfg.Pool, defaultTransactionTimeout, defaultBatchWriteSize), nil
	})
}
