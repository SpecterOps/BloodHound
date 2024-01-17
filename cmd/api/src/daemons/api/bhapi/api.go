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

package bhapi

import (
	"context"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/log"
)

// Daemon holds data relevant to the API daemon
type Daemon struct {
	cfg    config.Configuration
	server *http.Server
}

// NewDaemon creates a new API daemon
func NewDaemon(cfg config.Configuration, handler http.Handler) Daemon {
	networkTimeout := time.Duration(cfg.NetTimeoutSeconds) * time.Second

	return Daemon{
		cfg: cfg,
		server: &http.Server{
			Addr:         cfg.BindAddress,
			Handler:      handler,
			WriteTimeout: networkTimeout,
			ReadTimeout:  networkTimeout,
			IdleTimeout:  networkTimeout,
			ErrorLog:     log.Adapter(log.LevelError, "BHAPI", 0),
		},
	}
}

// Name returns the name of the daemon
func (s Daemon) Name() string {
	return "API Daemon"
}

// Start begins the daemon and waits for a stop signal in the exit channel
func (s Daemon) Start() {
	if s.cfg.TLS.Enabled() {
		if err := s.server.ListenAndServeTLS(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile); err != nil {
			if err != http.ErrServerClosed {
				log.Errorf("HTTP server listen error: %v", err)
			}
		}
	} else {
		if err := s.server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Errorf("HTTP server listen error: %v", err)
			}
		}
	}
}

// Stop passes in a stop signal to the exit channel, thereby killing the daemon
func (s Daemon) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
