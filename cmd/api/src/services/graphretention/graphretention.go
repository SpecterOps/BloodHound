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

// Empty implementation of graphretention adaptor
package graphretention

import (
	"context"
	"time"

	"github.com/specterops/bloodhound/packages/go/duration"
)

type GraphRetentionManager interface {
	GetBaseRetentionPolicyTTL(ctx context.Context) (duration.ISODuration, error)
	SetBaseRetentionPolicyTTL(ctx context.Context, ttl duration.ISODuration) error
}

type GraphRetentionAdapter struct{}

func (s GraphRetentionAdapter) GetBaseRetentionPolicyTTL(ctx context.Context) (duration.ISODuration, error) {
	return duration.ISODuration(time.Hour * 7 * 24), nil
}

func (s GraphRetentionAdapter) SetBaseRetentionPolicyTTL(ctx context.Context, ttl duration.ISODuration) error {
	return nil
}
