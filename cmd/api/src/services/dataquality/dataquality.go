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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/mock.go -package=mocks . DataQualityData
package dataquality

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/analysis/ad"
	"github.com/specterops/bloodhound/src/analysis/azure"
	"github.com/specterops/bloodhound/src/model"
)

type DataQualityData interface {
	CreateADDataQualityStats(stats model.ADDataQualityStats) (model.ADDataQualityStats, error)
	CreateADDataQualityAggregation(aggregation model.ADDataQualityAggregation) (model.ADDataQualityAggregation, error)
	CreateAzureDataQualityStats(stats model.AzureDataQualityStats) (model.AzureDataQualityStats, error)
	CreateAzureDataQualityAggregation(aggregation model.AzureDataQualityAggregation) (model.AzureDataQualityAggregation, error)
}

func SaveDataQuality(ctx context.Context, db DataQualityData, graphDB graph.Database) error {
	log.Infof("Started Data Quality Stats Collection")
	defer log.Measure(log.LevelInfo, "Successfully Completed Data Quality Stats Collection")()

	if stats, aggregation, err := ad.GraphStats(ctx, graphDB); err != nil {
		return fmt.Errorf("could not get active directory data quality stats: %w", err)
	} else if len(stats) > 0 {
		// We only want to save stats if there are stats to save
		if _, err := db.CreateADDataQualityStats(stats); err != nil {
			return fmt.Errorf("could not save active directory data quality stats: %w", err)
		} else if _, err := db.CreateADDataQualityAggregation(aggregation); err != nil {
			return fmt.Errorf("could not save active directory data quality aggregation: %w", err)
		}
	}

	if stats, aggregation, err := azure.GraphStats(ctx, graphDB); err != nil {
		return fmt.Errorf("could not get azure data quality stats: %w", err)
	} else if len(stats) > 0 {
		// We only want to save stats if there are stats to save
		if _, err := db.CreateAzureDataQualityStats(stats); err != nil {
			return fmt.Errorf("could not save azure data quality stats: %w", err)
		} else if _, err := db.CreateAzureDataQualityAggregation(aggregation); err != nil {
			return fmt.Errorf("could not save azure data quality stats: %w", err)
		}
	}

	return nil
}
