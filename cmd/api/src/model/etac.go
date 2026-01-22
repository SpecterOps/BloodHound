// Copyright 2025 Specter Ops, Inc.
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
package model

// EnvironmentTargetedAccessControl defines the model for a row in the environment_targeted_access_control table
type EnvironmentTargetedAccessControl struct {
	UserID        string `json:"user_id"`
	EnvironmentID string `json:"environment_id"`
	BigSerial
}

func (s EnvironmentTargetedAccessControl) Matches(x any) bool {
	mock, ok := x.(EnvironmentTargetedAccessControl)

	if !ok {
		return false
	} else if s.UserID != mock.UserID {
		return false
	} else if s.EnvironmentID != mock.UserID {
		return false
	}

	return true
}

func (EnvironmentTargetedAccessControl) TableName() string {
	return "environment_targeted_access_control"
}
