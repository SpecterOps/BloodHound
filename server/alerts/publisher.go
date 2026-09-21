// Copyright 2026 Specter Ops, Inc.
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

package alerts

import (
	"context"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
)

// AlertEventInput carries the values required to create a new alert event.
type AlertEventInput struct {
	Message string
	Data    types.JSONUntypedObject
}

type AlertEventType string

type Publisher interface {
	Publish(ctx context.Context, eventType AlertEventType, event AlertEventInput) error
}

type AlertEventPublisher struct{}

func NewAlertEventPublisher() *AlertEventPublisher {
	return &AlertEventPublisher{}
}

// Not implemented in BHCE
func (s *AlertEventPublisher) Publish(ctx context.Context, eventType AlertEventType, event AlertEventInput) error {
	return nil
}
