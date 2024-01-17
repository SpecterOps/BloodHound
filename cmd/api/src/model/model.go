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
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
)

// Basic is a struct which includes the following basic fields: CreatedAt, UpdatedAt, DeletedAt.
type Basic struct {
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" gorm:"-"`
}

// Unique is a struct is a struct which includes the following basic fields: ID, CreatedAt, UpdatedAt, DeletedAt.
type Unique struct {
	ID uuid.UUID `json:"id" gorm:"primaryKey"`

	Basic
}

// Serial is a struct which includes the following basic fields: ID, CreatedAt, UpdatedAt, DeletedAt.
// This was chosen over the default gorm model so that ID retains the bare int type. We do this because
// uint has no meaning with regards to the underlying database storage engine - at least where postgresql is
// concerned. To avoid type gnashing and unexpected pain with sql.NullInt32 the bare int type is a better
// choice all around.
//
// See: https://www.postgresql.org/docs/current/datatype-numeric.html
type Serial struct {
	ID int32 `json:"id" gorm:"primaryKey"`

	Basic
}

// BigSerial is a struct that follows the same design principles as Serial but with one exception:
// the ID type is set to int64 to support an ID sequence limit of up to 9223372036854775807.
type BigSerial struct {
	ID int64 `json:"id" gorm:"primaryKey"`

	Basic
}

// PaginatedResponse has been DEPRECATED as part of V1. Please use api.ResponseWrapper instead
type PaginatedResponse struct {
	Count int `json:"count"`
	Limit int `json:"limit"`
	Skip  int `json:"skip"`
	Data  any `json:"data"`
}

type BaseADEntity struct {
	Name     string `json:"name,omitempty"`
	ObjectID string `json:"objectid" gorm:"primaryKey"`
	Exists   bool   `json:"exists,omitempty"`
}

type LegacyADOU struct {
	BaseADEntity
	DistinguishedName string `json:"dn"`
	Type              string `json:"type"`
}

type PagedNodeListEntry struct {
	ObjectID string `json:"objectID"`
	Name     string `json:"name"`
	Label    string `json:"label"`
}

type DataType int

const (
	DataTypeGraph DataType = 0
	DataTypeList  DataType = 1
	DataTypeCount DataType = 2
)

type SearchResult struct {
	ObjectID          string `json:"objectid"`
	Type              string `json:"type"`
	Name              string `json:"name"`
	DistinguishedName string `json:"distinguishedname"`
	SystemTags        string `json:"system_tags"`
}
