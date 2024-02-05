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
	"github.com/specterops/bloodhound/log"

	"github.com/specterops/bloodhound/analysis"
	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/analysis/ad"
	"github.com/specterops/bloodhound/src/analysis/azure"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/services/agi"
	"github.com/specterops/bloodhound/src/services/dataquality"
)

var (
	ErrAnalysisFailed             = errors.New("analysis failed")
	ErrAnalysisPartiallyCompleted = errors.New("analysis partially completed")
)

func RunAnalysisOperations(ctx context.Context, db database.Database, graphDB graph.Database, _ config.Configuration) error {
	var (
		collectedErrors []error
	)

	if err := adAnalysis.FixWellKnownNodeTypes(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("fix well known node types failed: %w", err))
	}

	if err := adAnalysis.RunDomainAssociations(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("domain association and pruning failed: %w", err))
	}

	if err := adAnalysis.LinkWellKnownGroups(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("well known group linking failed: %w", err))
	}

	if err := updateAssetGroupIsolationTags(ctx, db, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("asset group isolation tagging failed: %w", err))
	}

	if err := TagActiveDirectoryTierZero(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("active directory tier zero tagging failed: %w", err))
	}

	if err := ParallelTagAzureTierZero(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("azure tier zero tagging failed: %w", err))
	}

	var (
		adFailed          = false
		azureFailed       = false
		agiFailed         = false
		dataQualityFailed = false
	)

	// TODO: Cleanup #ADCSFeatureFlag after full launch.
	if adcsFlag, err := db.GetFlagByKey(appcfg.FeatureAdcs); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error retrieving ADCS feature flag: %w", err))
	} else if stats, err := ad.Post(ctx, graphDB, adcsFlag.Enabled); err != nil {
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

	if err := agi.RunAssetGroupIsolationCollections(ctx, db, graphDB, analysis.GetNodeKindDisplayLabel); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("asset group isolation collection failed: %w", err))
		agiFailed = true
	}

	if err := dataquality.SaveDataQuality(ctx, db, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error saving data quality stat: %v", err))
		dataQualityFailed = true
	}

	if len(collectedErrors) > 0 {
		for _, err := range collectedErrors {
			log.Errorf("Analysis error encountered: %v", err)
		}
	}

	if adFailed && azureFailed && agiFailed && dataQualityFailed {
		return ErrAnalysisFailed
	} else if adFailed || azureFailed || agiFailed || dataQualityFailed {
		return ErrAnalysisPartiallyCompleted
	}

	return nil
}
