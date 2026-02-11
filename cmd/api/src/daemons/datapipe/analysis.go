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

package datapipe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/cmd/api/src/analysis/ad"
	"github.com/specterops/bloodhound/cmd/api/src/analysis/azure"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/agi"
	"github.com/specterops/bloodhound/cmd/api/src/services/dataquality"
	"github.com/specterops/bloodhound/packages/go/analysis"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/dawgs/graph"
)

var (
	ErrAnalysisFailed             = errors.New("analysis failed")
	ErrAnalysisPartiallyCompleted = errors.New("analysis partially completed")
)

// TODO Cleanup tieringEnabled after Tiering GA
func RunAnalysisOperations(ctx context.Context, db database.Database, graphDB graph.Database, _ config.Configuration) error {
	var (
		collectedErrors      []error
		compositionIdCounter = analysis.NewCompositionCounter()
		tieringEnabled       = appcfg.GetTieringEnabled(ctx, db)
	)

	if err := adAnalysis.FixWellKnownNodeTypes(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("fix well known node types failed: %w", err))
	}

	if err := adAnalysis.RunDomainAssociations(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("domain association and pruning failed: %w", err))
	}

	if err := adAnalysis.LinkWellKnownNodes(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("well known group linking failed: %w", err))
	}

	var (
		adFailed          = false
		azureFailed       = false
		agiFailed         = false
		dataQualityFailed = false
	)

	// TODO: Cleanup #ADCSFeatureFlag after full launch.
	if adcsFlag, err := db.GetFlagByKey(ctx, appcfg.FeatureAdcs); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error retrieving ADCS feature flag: %w", err))
	} else if ntlmFlag, err := db.GetFlagByKey(ctx, appcfg.FeatureNTLMPostProcessing); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error retrieving NTLM Post Processing feature flag: %w", err))
	} else if stats, err := ad.Post(ctx, graphDB, adcsFlag.Enabled, appcfg.GetCitrixRDPSupport(ctx, db), ntlmFlag.Enabled, &compositionIdCounter); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error during ad post: %w", err))
		adFailed = true
	} else {
		stats.LogStats()
	}

	if stats, err := azure.Post(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error during azure post: %w", err))
		azureFailed = true
	} else {
		stats.LogStats()
	}

	if errs := TagAssetGroupsAndTierZero(ctx, db, graphDB); len(errs) > 0 {
		for _, err := range errs {
			collectedErrors = append(collectedErrors, fmt.Errorf("tagging asset groups and tier zero failed: %w", err))
		}
	}

	if !tieringEnabled {
		if err := agi.RunAssetGroupIsolationCollections(ctx, db, graphDB); err != nil {
			collectedErrors = append(collectedErrors, fmt.Errorf("asset group isolation collection failed: %w", err))
			agiFailed = true
		}
	}

	if err := dataquality.SaveDataQuality(ctx, db, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error saving data quality stat: %v", err))
		dataQualityFailed = true
	}

	if len(collectedErrors) > 0 {
		for _, err := range collectedErrors {
			slog.ErrorContext(ctx, fmt.Sprintf("Analysis error encountered: %v", err))
		}
	}

	if adFailed && azureFailed && agiFailed && dataQualityFailed {
		return ErrAnalysisFailed
	} else if adFailed || azureFailed || agiFailed || dataQualityFailed {
		return ErrAnalysisPartiallyCompleted
	}

	return nil
}
