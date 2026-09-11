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

// This file defines the external interface methods allowing
// access to specific parameter types' values
package services

import (
	"time"
)

const (
	PasswordExpirationWindow ParameterKey = "auth.password_expiration_window"
	SessionTTLHours          ParameterKey = "auth.session_ttl_hours"
	Neo4jConfigs             ParameterKey = "neo4j.configuration"
	CitrixRDPSupportKey      ParameterKey = "analysis.citrix_rdp_support"
	PruneTTL                 ParameterKey = "prune.ttl"
	ReconciliationKey        ParameterKey = "analysis.reconciliation"
	ScheduledAnalysis        ParameterKey = "analysis.scheduled"
	ClientMetricsKey         ParameterKey = "pipeline.client_metrics"
	APITokenExpiration       ParameterKey = "auth.api_token_expiration"

	// The below keys are not intended to be user updatable, so should not be added to IsValidKey
	TrustedProxiesConfig                ParameterKey = "http.trusted_proxies"
	FedEULACustomTextKey                ParameterKey = "eula.custom_text"
	TierManagementParameterKey          ParameterKey = "analysis.tiering"
	AGTParameterKey                     ParameterKey = "analysis.tagging"
	StaleClientUpdatedLogicKey          ParameterKey = "pipeline.updated_stale_client"
	RetainIngestedFilesKey              ParameterKey = "analysis.retain_ingest_files"
	APITokens                           ParameterKey = "auth.api_tokens"
	TimeoutLimit                        ParameterKey = "api.timeout_limit"
	EnvironmentTargetedAccessControlKey ParameterKey = "auth.environment_targeted_access_control"
	SupportAccountProvisioningKey       ParameterKey = "auth.support_account_provisioning"
	GraphStorageOptimizationKey         ParameterKey = "analysis.graph_storage_optimization"
)

const (
	DefaultPasswordExpirationWindow = time.Hour * 24 * 90

	DefaultSessionTTLHours = 8

	DefaultPruneBaseTTL           = time.Hour * 24 * 7
	DefaultPruneHasSessionEdgeTTL = time.Hour * 24 * 3

	MaxDawgsWorkerLimit         = 6 // This is the maximum analysis parallel workers during tagging
	DefaultDawgsWorkerLimit     = 2 // This is the parallel workers during tagging
	DefaultExpansionWorkerLimit = 3 // This is the size of the expansion worker pool during tagging
	DefaultSelectorWorkerLimit  = 7 // This is the size of the selector worker pool during tagging
)
