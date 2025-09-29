// Copyright 2025 Specter Ops, Inc.
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
package ha

import "context"

type LockResult struct {
	Context   context.Context
	IsPrimary bool
}

type HAMutex interface {
	TryLock() (LockResult, error)
}

// dummyHA is a no-op implementation for BHCE that always reports as primary
type dummyHA struct{}

func (d *dummyHA) TryLock() (LockResult, error) {
	return LockResult{
		Context:   context.Background(),
		IsPrimary: true,
	}, nil
}

func NewDummyHA() HAMutex {
	return &dummyHA{}
}
