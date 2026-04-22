// Copyright 2026 Specter Ops, Inc.
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

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
)

type Webhooks []Webhook

type Webhook struct {
	ID          uuid.UUID   `json:"id" gorm:"type:text;primaryKey"`
	Type        string      `json:"type" validate:"required"`
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description"`
	URL         string      `json:"url" validate:"required"`
	CreatedAt   time.Time   `json:"created_at"`
	CreatedBy   string      `json:"created_by"`
	UpdatedAt   time.Time   `json:"updated_at"`
	UpdatedBy   string      `json:"updated_by"`
	DisabledAt  null.Time   `json:"disabled_at"`
	DisabledBy  null.String `json:"disabled_by"`
}

func (Webhook) TableName() string {
	return "webhooks"
}

func (s Webhook) AuditData() AuditData {
	return AuditData{
		"id":          s.ID,
		"type":        s.Type,
		"name":        s.Name,
		"description": s.Description,
		"url":         s.URL,
		"created_at":  s.CreatedAt,
		"created_by":  s.CreatedBy,
		"updated_at":  s.UpdatedAt,
		"updated_by":  s.UpdatedBy,
		"disabled_at": s.DisabledAt,
		"disabled_by": s.DisabledBy,
	}
}

func (s Webhook) IsStringColumn(filter string) bool {
	switch filter {
	case "type", "name", "description", "url", "created_by", "updated_by", "disabled_by":
		return true
	default:
		return false
	}
}

func (s Webhook) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"type":        {Equals, NotEquals, ApproximatelyEquals},
		"name":        {Equals, NotEquals, ApproximatelyEquals},
		"description": {Equals, NotEquals, ApproximatelyEquals},
		"url":         {Equals, NotEquals, ApproximatelyEquals},
		"created_at":  {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_by":  {Equals, NotEquals},
		"updated_at":  {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_by":  {Equals, NotEquals},
		"disabled_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"disabled_by": {Equals, NotEquals},
	}
}

func (s Webhooks) IsSortable(column string) bool {
	switch column {
	case "id", "type", "name", "url", "created_at", "updated_at", "disabled_at":
		return true
	default:
		return false
	}
}

type WebhookSecret struct {
	WebhookId  uuid.UUID `json:"webhook_id" gorm:"type:text;primaryKey"`
	HmacSecret string    `json:"hmac_secret"`
	CreatedAt  time.Time `json:"created_at"`
}

func (WebhookSecret) TableName() string {
	return "webhook_secrets"
}

type WebhookMetadata struct {
	WebhookId       uuid.UUID   `json:"webhook_id" gorm:"type:text;primaryKey"`
	Health          float64     `json:"health"`
	Attempts        int         `json:"attempts"`
	Failures        int         `json:"failures"`
	LastError       null.String `json:"last_error"`
	LastErroredAt   null.Time   `json:"last_errored_at"`
	LastSucceededAt null.Time   `json:"last_succeeded_at"`
}

func (WebhookMetadata) TableName() string {
	return "webhook_metadata"
}

type WebhookSubscriptions []WebhookSubscription

type WebhookSubscription struct {
	WebhookId uuid.UUID `json:"webhook_id" gorm:"type:text;primaryKey"`
	EventType string    `json:"event_type" gorm:"primaryKey" validate:"required"`
	Version   string    `json:"version" validate:"required"`
}

func (WebhookSubscription) TableName() string {
	return "webhook_subscriptions"
}

func (s WebhookSubscription) AuditData() AuditData {
	return AuditData{
		"webhook_id": s.WebhookId,
		"event_type": s.EventType,
		"version":    s.Version,
	}
}

type WebhookEvents []WebhookEvent

type WebhookEvent struct {
	WebhookId      uuid.UUID   `json:"webhook_id" gorm:"type:text;primaryKey"`
	EventId        uuid.UUID   `json:"event_id" gorm:"type:text;primaryKey"`
	CreatedAt      time.Time   `json:"created_at"`
	LastStatusCode null.Int32  `json:"last_status_code"`
	LastError      null.String `json:"last_error"`
	Attempts       int         `json:"attempts"`
}

func (WebhookEvent) TableName() string {
	return "webhook_events"
}
