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

type AzureDataQualityStat struct {
	TenantID          string `json:"tenantid" gorm:"column:tenant_id"`
	Users             int    `json:"users"`
	Groups            int    `json:"groups"`
	Apps              int    `json:"apps"`
	ServicePrincipals int    `json:"service_principals"`
	Devices           int    `json:"devices"`
	ManagementGroups  int    `json:"management_groups"`
	Subscriptions     int    `json:"subscriptions"`
	ResourceGroups    int    `json:"resource_groups"`
	VMs               int    `json:"vms"`
	KeyVaults         int    `json:"key_vaults"`
	Relationships     int    `json:"relationships"`
	RunID             string `json:"run_id" gorm:"index"`

	Serial
}

type AzureDataQualityAggregation struct {
	Tenants           int    `json:"tenants"`
	Users             int    `json:"users"`
	Groups            int    `json:"groups"`
	Apps              int    `json:"apps"`
	ServicePrincipals int    `json:"service_principals"`
	Devices           int    `json:"devices"`
	ManagementGroups  int    `json:"management_groups"`
	Subscriptions     int    `json:"subscriptions"`
	ResourceGroups    int    `json:"resource_groups"`
	VMs               int    `json:"vms"`
	KeyVaults         int    `json:"key_vaults"`
	Relationships     int    `json:"relationships"`
	RunID             string `json:"run_id" gorm:"index"`

	Serial
}

type AzureDataQualityStats []AzureDataQualityStat

type AzureDataQualityAggregations []AzureDataQualityAggregation

func (s AzureDataQualityStats) IsSortable(column string) bool {
	switch column {
	case "updated_at",
		"created_at":
		return true
	default:
		return false
	}
}
