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

//go:build integration
// +build integration

package ad_test

import (
	"context"
	"github.com/specterops/bloodhound/analysis"
	ad2 "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestADCSESC1(t *testing.T) {
	testContext := integration.NewGraphTestContext(t)
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) {
		harness.ADCSESC1Harness.Setup(testContext)
	}, func(harness integration.HarnessDetails, db graph.Database) error {
		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCAs, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		enrollCache, err := ad2.BuildEsc1Cache(context.Background(), db, enterpriseCAs, certTemplates)

		var results cardinality.Duplex[uint32]
		err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			innerResults, err := ad2.PostADCSESC1Domain(tx, harness.ADCSESC1Harness.Domain1, groupExpansions, enrollCache)
			results = innerResults
			return err
		})
		require.Nil(t, err)

		require.True(t, results.Cardinality() == 3)
		require.True(t, results.Contains(harness.ADCSESC1Harness.User13.ID.Uint32()))
		require.True(t, results.Contains(harness.ADCSESC1Harness.User11.ID.Uint32()))
		require.True(t, results.Contains(harness.ADCSESC1Harness.Group13.ID.Uint32()))

		err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			innerResults, err := ad2.PostADCSESC1Domain(tx, harness.ADCSESC1Harness.Domain2, groupExpansions, enrollCache)
			results = innerResults
			return err
		})

		require.Nil(t, err)
		require.True(t, results.Cardinality() == 1)
		require.True(t, results.Contains(harness.ADCSESC1Harness.Group22.ID.Uint32()))

		err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			innerResults, err := ad2.PostADCSESC1Domain(tx, harness.ADCSESC1Harness.Domain3, groupExpansions, enrollCache)
			results = innerResults
			return err
		})

		require.Nil(t, err)
		require.True(t, results.Cardinality() == 1)
		require.True(t, results.Contains(harness.ADCSESC1Harness.Group32.ID.Uint32()))

		err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			innerResults, err := ad2.PostADCSESC1Domain(tx, harness.ADCSESC1Harness.Domain4, groupExpansions, enrollCache)
			results = innerResults
			return err
		})

		require.Nil(t, err)
		require.True(t, results.Cardinality() == 2)
		require.True(t, results.Contains(harness.ADCSESC1Harness.Group42.ID.Uint32()))
		require.True(t, results.Contains(harness.ADCSESC1Harness.Group43.ID.Uint32()))

		return nil
	})
}

func TestEnrollOnBehalfOf(t *testing.T) {
	testContext := integration.NewGraphTestContext(t)
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) {
		harness.EnrollOnBehalfOfHarnessOne.Setup(testContext)
	}, func(harness integration.HarnessDetails, db graph.Database) error {
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		v1Templates := make([]*graph.Node, 0)
		v2Templates := make([]*graph.Node, 0)
		for _, template := range certTemplates {
			if version, err := template.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
				continue
			} else if version == 1 {
				v1Templates = append(v1Templates, template)
			} else if version >= 2 {
				v2Templates = append(v2Templates, template)
			}
		}
		require.Nil(t, err)
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			results, err := ad2.EnrollOnBehalfOfVersionOne(tx, v1Templates, certTemplates)
			require.Nil(t, err)

			require.Len(t, results, 2)

			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarnessOne.CertTemplate11.ID,
				ToID:   harness.EnrollOnBehalfOfHarnessOne.CertTemplate12.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarnessOne.CertTemplate13.ID,
				ToID:   harness.EnrollOnBehalfOfHarnessOne.CertTemplate12.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			return nil
		})

		return nil
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) {
		harness.EnrollOnBehalfOfHarnessTwo.Setup(testContext)
	}, func(harness integration.HarnessDetails, db graph.Database) error {
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		v1Templates := make([]*graph.Node, 0)
		v2Templates := make([]*graph.Node, 0)
		for _, template := range certTemplates {
			if version, err := template.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
				continue
			} else if version == 1 {
				v1Templates = append(v1Templates, template)
			} else if version >= 2 {
				v2Templates = append(v2Templates, template)
			}
		}
		require.Nil(t, err)
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			results, err := ad2.EnrollOnBehalfOfVersionTwo(tx, v2Templates, certTemplates)
			require.Nil(t, err)

			require.Len(t, results, 1)
			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarnessTwo.CertTemplate21.ID,
				ToID:   harness.EnrollOnBehalfOfHarnessTwo.CertTemplate23.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			results, err = ad2.EnrollOnBehalfOfSelfControl(tx, v1Templates)
			require.Nil(t, err)

			require.Len(t, results, 1)
			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarnessTwo.CertTemplate25.ID,
				ToID:   harness.EnrollOnBehalfOfHarnessTwo.CertTemplate25.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			return nil
		})

		return nil
	})
}
