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

package gc

import (
	"context"
	"time"

	"github.com/specterops/bloodhound/src/database"
)

// Daemon holds data relevant to the data daemon
type Daemon struct {
	exitC chan struct{}
	db    database.Database
}

// NewDataPruningDaemon creates a new data pruning daemon
func NewDataPruningDaemon(db database.Database) *Daemon {
	return &Daemon{
		exitC: make(chan struct{}),
		db:    db,
	}
}

// Name returns the name of the daemon
func (s *Daemon) Name() string {
	return "Data Pruning Daemon"
}

// Start begins the daemon and waits for a stop signal in the exit channel
func (s *Daemon) Start() {
	ticker := time.NewTicker(24 * time.Hour)

	defer close(s.exitC)
	defer ticker.Stop()

	// prune sessions and collections once when the daemon starts up
	s.db.SweepSessions()
	s.db.SweepAssetGroupCollections()

	// thereafter, prune conditionally once a day
	for {
		select {
		case <-ticker.C:
			s.db.SweepSessions()
			s.db.SweepAssetGroupCollections()

		case <-s.exitC:
			return
		}
	}
}

// Stop passes in a stop signal to the exit channel, thereby killing the daemon
func (s *Daemon) Stop(ctx context.Context) error {
	s.exitC <- struct{}{}

	select {
	case <-s.exitC:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
