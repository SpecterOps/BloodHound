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

package config

import (
	"fmt"
	"time"

	"github.com/specterops/bloodhound/crypto"
)

type Argon2Configuration struct {
	MemoryKibibytes uint32 `json:"memory_kibibytes"`
	NumIterations   uint32 `json:"num_iterations"`
	NumThreads      uint8  `json:"num_threads"`
}

func GenerateArgonSettings(tuneDuration time.Duration, skipArgon2 bool) (Argon2Configuration, error) {
	var (
		digester crypto.Argon2
		err      error
	)

	if skipArgon2 {
		return Argon2Configuration{}, nil
	} else if digester, err = crypto.Tune(time.Millisecond * tuneDuration); err != nil {
		return Argon2Configuration{}, fmt.Errorf("failed tuning argon2: %w", err)
	} else {
		return Argon2Configuration{
			MemoryKibibytes: digester.MemoryKibibytes,
			NumIterations:   digester.NumIterations,
			NumThreads:      digester.NumThreads,
		}, nil
	}
}
