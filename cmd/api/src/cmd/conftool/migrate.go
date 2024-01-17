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

package main

import (
	"github.com/specterops/bloodhound/src/config"
)

func migrateConfiguration(path string) (config.Configuration, error) {
	if lastConfig, err := config.ReadConfigurationFile(path); err != nil {
		return lastConfig, err
	} else {
		for lastConfig.Version != config.CurrentConfigurationVersion {
			switch lastConfig.Version {
			case 0:
				// Currently a no-op
			case 1:
				// Currently a no-op
			}

			lastConfig.Version += 1
		}

		return lastConfig, nil
	}
}
