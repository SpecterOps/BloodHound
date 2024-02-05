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

import "github.com/specterops/bloodhound/src/model"

const (
	FeatureButterflyAnalysis   = "butterfly_analysis"
	FeatureEnableSAMLSSO       = "enable_saml_sso"
	FeatureScopeCollectionByOU = "scope_collection_by_ou"
	FeatureAzureSupport        = "azure_support"
	FeatureReconciliation      = "reconciliation"
	FeatureEntityPanelCaching  = "entity_panel_cache"
	FeatureAdcs                = "adcs"
)

// AvailableFlags returns a FeatureFlagSet of expected feature flags. Feature flag defaults introduced here will become the initial
// default value of the feature flag once it is inserted into the database.
func AvailableFlags() FeatureFlagSet {
	return FeatureFlagSet{
		FeatureButterflyAnalysis: {
			Key:           FeatureButterflyAnalysis,
			Name:          "Enhanced Asset Inbound-Outbound Exposure Analysis",
			Description:   "Enables more extensive analysis of attack path findings that allows BloodHound to help the user prioritize remediation of the most exposed assets.",
			Enabled:       false,
			UserUpdatable: false,
		},
		FeatureEnableSAMLSSO: {
			Key:           FeatureEnableSAMLSSO,
			Name:          "SAML Single Sign-On Support",
			Description:   "Enables SSO authentication flows and administration panels to third party SAML identity providers.",
			Enabled:       true,
			UserUpdatable: false,
		},
		FeatureScopeCollectionByOU: {
			Key:           FeatureScopeCollectionByOU,
			Name:          "Enable SharpHound OU Scoped Collections",
			Description:   "Enables scoping SharpHound collections to specific lists of OUs.",
			Enabled:       true,
			UserUpdatable: false,
		},
		FeatureAzureSupport: {
			Key:           FeatureAzureSupport,
			Name:          "Enable Azure Support",
			Description:   "Enables Azure support.",
			Enabled:       true,
			UserUpdatable: false,
		},
		FeatureReconciliation: {
			Key:           FeatureReconciliation,
			Name:          "Reconciliation",
			Description:   "Enables Reconciliation",
			Enabled:       true,
			UserUpdatable: false,
		},
		FeatureEntityPanelCaching: {
			Key:           FeatureEntityPanelCaching,
			Name:          "Enable application level caching",
			Description:   "Enables the use of application level caching for entity panel queries",
			Enabled:       true,
			UserUpdatable: false,
		},
		FeatureAdcs: {
			Key:           FeatureAdcs,
			Name:          "Enable collection and processing of Active Directory Certificate Services Data",
			Description:   "Enables the ability to collect, analyze, and explore Active Directory Certificate Services data and previews new attack paths.",
			Enabled:       false,
			UserUpdatable: false,
		},
	}
}

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
	// GetAllFlags gets all available runtime feature flags as a FeatureFlagSet for the application.
	GetAllFlags() ([]FeatureFlag, error)

	// GetFlag attempts to fetch a FeatureFlag by its ID.
	GetFlag(id int32) (FeatureFlag, error)

	// GetFlagByKey attempts to fetch a FeatureFlag by its key.
	GetFlagByKey(key string) (FeatureFlag, error)

	// SetFlag attempts to store or update the given FeatureFlag by its feature Key.
	SetFlag(value FeatureFlag) error
}
