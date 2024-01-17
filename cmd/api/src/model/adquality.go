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
	"github.com/specterops/bloodhound/src/database/types/nan"
)

type ADDataQualityStat struct {
	DomainSID              string      `json:"domain_sid" gorm:"column:domain_sid"`
	Users                  int         `json:"users"`
	Groups                 int         `json:"groups"`
	Computers              int         `json:"computers"`
	OUs                    int         `json:"ous" gorm:"column:ous"`
	Containers             int         `json:"containers"`
	GPOs                   int         `json:"gpos" gorm:"column:gpos"`
	AIACAs                 int         `json:"aiacas" gorm:"column:aiacas"`
	RootCAs                int         `json:"rootcas" gorm:"column:rootcas"`
	EnterpriseCAs          int         `json:"enterprisecas" gorm:"column:enterprisecas"`
	NTAuthStores           int         `json:"ntauthstores" gorm:"column:ntauthstores"`
	CertTemplates          int         `json:"certtemplates" gorm:"column:certtemplates"`
	ACLs                   int         `json:"acls" gorm:"column:acls"`
	Sessions               int         `json:"sessions"`
	Relationships          int         `json:"relationships"`
	SessionCompleteness    nan.Float64 `json:"session_completeness"`
	LocalGroupCompleteness nan.Float64 `json:"local_group_completeness"`
	RunID                  string      `json:"run_id" gorm:"index"`

	Serial
}

type ADDataQualityAggregation struct {
	Domains                int     `json:"domains"`
	Users                  int     `json:"users"`
	Groups                 int     `json:"groups"`
	Computers              int     `json:"computers"`
	OUs                    int     `json:"ous" gorm:"column:ous"`
	Containers             int     `json:"containers"`
	GPOs                   int     `json:"gpos" gorm:"column:gpos"`
	AIACAs                 int     `json:"aiacas" gorm:"column:aiacas"`
	RootCAs                int     `json:"rootcas" gorm:"column:rootcas"`
	EnterpriseCAs          int     `json:"enterprisecas" gorm:"column:enterprisecas"`
	NTAuthStores           int     `json:"ntauthstores" gorm:"column:ntauthstores"`
	CertTemplates          int     `json:"certtemplates" gorm:"column:certtemplates"`
	Acls                   int     `json:"acls" gorm:"column:acls"`
	Sessions               int     `json:"sessions"`
	Relationships          int     `json:"relationships"`
	SessionCompleteness    float32 `json:"session_completeness"`
	LocalGroupCompleteness float32 `json:"local_group_completeness"`
	RunID                  string  `json:"run_id" gorm:"index"`

	Serial
}

type ADDataQualityStats []ADDataQualityStat

type ADDataQualityAggregations []ADDataQualityAggregation

func (s ADDataQualityStats) IsSortable(column string) bool {
	switch column {
	case "updated_at",
		"created_at":
		return true
	default:
		return false
	}
}
