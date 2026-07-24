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

package alertevents

import "context"

// Publisher records an alert event for downstream dispatch.
// BHE only so the default Publisher is a no-op.
type Publisher interface {
	Publish(ctx context.Context, eventType, message string, data map[string]any) error
}

// NoopPublisher discards all events. Used wherever the alerts subsystem is
// unavailable (BHCE standalone).
type NoopPublisher struct{}

func NewNoopPublisher() NoopPublisher { return NoopPublisher{} }

func (s NoopPublisher) Publish(context.Context, string, string, map[string]any) error {
	return nil
}
