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

package model

import (
	"time"

	"github.com/specterops/bloodhound/src/version"
)

type Migration struct {
	ID        int32 `gorm:"primaryKey"`
	UpdatedAt time.Time
	Major     int32
	Minor     int32
	Patch     int32
}

func NewMigration(target version.Version) Migration {
	return Migration{
		UpdatedAt: time.Now(),
		Major:     int32(target.Major),
		Minor:     int32(target.Minor),
		Patch:     int32(target.Patch),
	}
}

func (s Migration) Version() version.Version {
	return version.Version{
		Major: int(s.Major),
		Minor: int(s.Minor),
		Patch: int(s.Patch),
	}
}
