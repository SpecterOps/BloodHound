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

package appcfg

import (
	"context"
	"log/slog"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// AvailableFlags has been removed and the db feature_flags table is the source of truth. Feature flag defaults should be added via migration *.sql files.
const (
	FeatureButterflyAnalysis            = "butterfly_analysis"
	FeatureEnableSAMLSSO                = "enable_saml_sso"
	FeatureScopeCollectionByOU          = "scope_collection_by_ou"
	FeatureAzureSupport                 = "azure_support"
	FeatureEntityPanelCaching           = "entity_panel_cache"
	FeatureAdcs                         = "adcs"
	FeatureClearGraphData               = "clear_graph_data"
	FeatureRiskExposureNewCalculation   = "risk_exposure_new_calculation"
	FeatureFedRAMPEULA                  = "fedramp_eula"
	FeatureDarkMode                     = "dark_mode"
	FeatureAutoTagT0ParentObjects       = "auto_tag_t0_parent_objects"
	FeatureOIDCSupport                  = "oidc_support"
	FeatureNTLMPostProcessing           = "ntlm_post_processing"
	FeatureTierManagement               = "tier_management_engine"
	FeatureChangelog                    = "changelog"
	FeatureETAC                         = "environment_targeted_access_control"
	FeatureOpenGraphSearch              = "opengraph_search"
	FeatureOpenGraphFindings            = "opengraph_findings"
	FeatureClientBearerAuth             = "client_bearer_auth"
	FeatureOpenGraphExtensionManagement = "opengraph_extension_management"
)

// FeatureFlag defines the most basic details of what a feature flag must contain to be actionable. Feature flags should be
// self-descriptive as many use-cases will involve iterating over all available flags to display them back to the
// end-user.
type FeatureFlag struct {
	model.Serial

	// Key is the unique identifier for this feature flag that is also used as its storage-key. This is intended only
	// for internal referencing to and from the API when scoping operations to just this feature flag.
	Key string `json:"key" gorm:"unique"`

	// Name is a display friendly name for this particular flag.
	Name string `json:"name"`

	// Description is a display friendly paragraph describing the intent and utilization of the feature flag.
	Description string `json:"description"`

	// Enabled determines if the feature flag is active or not.
	Enabled bool `json:"enabled"`

	// UserUpdatable determines whether a user with the correct permissions can change the enablement of this feature flag.
	// Note that this does not prevent the system, in-code, from modifying the feature flag's state. The scope of this
	// value only applies to user interaction flows.
	UserUpdatable bool `json:"user_updatable"`
}

// FeatureFlagSet is a collection of flags indexed by their flag Key.
type FeatureFlagSet map[string]FeatureFlag

// FeatureFlagService defines a contract for fetching and setting feature flags.
type FeatureFlagService interface {
	GetFlagByKeyer

	// GetAllFlags gets all available runtime feature flags as a FeatureFlagSet for the application.
	GetAllFlags(ctx context.Context) ([]FeatureFlag, error)

	// GetFlag attempts to fetch a FeatureFlag by its ID.
	GetFlag(ctx context.Context, id int32) (FeatureFlag, error)

	// SetFlag attempts to store or update the given FeatureFlag by its feature Key.
	SetFlag(ctx context.Context, value FeatureFlag) error
}

type GetFlagByKeyer interface {
	// GetFlagByKey attempts to fetch a FeatureFlag by its key.
	GetFlagByKey(context.Context, string) (FeatureFlag, error)
}

// TODO Cleanup after Tiering GA
func GetTieringEnabled(ctx context.Context, service GetFlagByKeyer) bool {
	if tierFlag, err := service.GetFlagByKey(ctx, FeatureTierManagement); err != nil {
		slog.WarnContext(ctx, "Failed to fetch tiering management flag; returning false")
		return false
	} else {
		return tierFlag.Enabled
	}
}

func (s FeatureFlag) AuditData() model.AuditData {
	return model.AuditData{
		"name":    s.Name,
		"key":     s.Key,
		"enabled": s.Enabled,
	}
}
