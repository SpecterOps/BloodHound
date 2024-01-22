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

package dawgs

import (
	"context"
	"errors"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

var (
	ErrDriverMissing = errors.New("driver missing")
)

type DriverConstructor func(ctx context.Context, cfg Config) (graph.Database, error)

var availableDrivers = map[string]DriverConstructor{}

func Register(driverName string, constructor DriverConstructor) {
	availableDrivers[driverName] = constructor
}

type Config struct {
	TraversalMemoryLimit size.Size
	DriverCfg            any
}

func Open(ctx context.Context, driverName string, config Config) (graph.Database, error) {
	if driverConstructor, hasDriver := availableDrivers[driverName]; !hasDriver {
		return nil, ErrDriverMissing
	} else {
		return driverConstructor(ctx, config)
	}
}
