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

package daemons

import (
	"context"
	"sync"
	"time"

	"github.com/specterops/bloodhound/log"
)

type Daemon interface {
	Start()
	Name() string
	Stop(ctx context.Context) error
}

type Manager struct {
	daemons         []Daemon
	daemonsLock     *sync.Mutex
	shutdownTimeout time.Duration
}

func NewManager(shutdownTimeout time.Duration) *Manager {
	return &Manager{
		daemonsLock:     &sync.Mutex{},
		shutdownTimeout: shutdownTimeout,
	}
}

func (s *Manager) Start(daemons ...Daemon) {
	s.daemonsLock.Lock()
	defer s.daemonsLock.Unlock()

	for _, daemon := range daemons {
		log.Infof("Starting daemon %s", daemon.Name())
		go daemon.Start()

		s.daemons = append(s.daemons, daemon)
	}
}

func (s *Manager) Stop() {
	s.daemonsLock.Lock()
	defer s.daemonsLock.Unlock()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	for _, daemon := range s.daemons {
		log.Infof("Shutting down daemon %s", daemon.Name())

		if err := daemon.Stop(shutdownCtx); err != nil {
			log.Errorf("Failure caught while shutting down daemon %s: %v", daemon.Name(), err)
		}
	}
}
