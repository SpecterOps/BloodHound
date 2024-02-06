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

package ad

import (
	"context"
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

var (
	ErrNoCertParent     = errors.New("cert has no parent")
	EkuAnyPurpose       = "2.5.29.37.0"
	EkuCertRequestAgent = "1.3.6.1.4.1.311.20.2.1"
)

func PostADCS(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator, adcsEnabled bool) (*analysis.AtomicPostProcessingStats, error) {
	if enterpriseCertAuthorities, err := FetchNodesByKind(ctx, db, ad.EnterpriseCA); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching enterpriseCA nodes: %w", err)
	} else if rootCertAuthorities, err := FetchNodesByKind(ctx, db, ad.RootCA); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching rootCA nodes: %w", err)
	} else if certTemplates, err := FetchNodesByKind(ctx, db, ad.CertTemplate); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching cert template nodes: %w", err)
	} else if domains, err := FetchNodesByKind(ctx, db, ad.Domain); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching domain nodes: %w", err)
	} else if step1Stats, err := postADCSPreProcessStep1(ctx, db, enterpriseCertAuthorities, rootCertAuthorities); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed adcs pre-processing step 1: %w", err)
	} else if step2Stats, err := postADCSPreProcessStep2(ctx, db, certTemplates); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed adcs pre-processing step 2: %w", err)
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing")

		operation.Stats.Merge(step1Stats)
		operation.Stats.Merge(step2Stats)

		var cache = NewADCSCache()
		cache.BuildCache(ctx, db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {
					processEnterpriseCAWithValidCertChainToDomain(innerEnterpriseCA, innerDomain, groupExpansions, cache, operation, adcsEnabled)
				}
			}
		}

		return &operation.Stats, operation.Done()
	}
}

// postADCSPreProcessStep1 processes the edges that are not dependent on any other post-processed edges
func postADCSPreProcessStep1(ctx context.Context, db graph.Database, enterpriseCertAuthorities, rootCertAuthorities []*graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 1")
	// TODO clean up the operation.Done() calls below

	if err := PostTrustedForNTAuth(ctx, db, operation); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.TrustedForNTAuth.String(), err)
	} else if err := PostIssuedSignedBy(operation, enterpriseCertAuthorities, rootCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.IssuedSignedBy.String(), err)
	} else if err := PostEnterpriseCAFor(operation, enterpriseCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnterpriseCAFor.String(), err)
	} else if err = PostCanAbuseUPNCertMapping(operation, enterpriseCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.CanAbuseUPNCertMapping.String(), err)
	} else if err = PostCanAbuseWeakCertBinding(operation, enterpriseCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.CanAbuseWeakCertBinding.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

// postADCSPreProcessStep2 Processes the edges that are dependent on those processed in postADCSPreProcessStep1
func postADCSPreProcessStep2(ctx context.Context, db graph.Database, certTemplates []*graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 2")

	if err := PostEnrollOnBehalfOf(certTemplates, operation); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnrollOnBehalfOf.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

func processEnterpriseCAWithValidCertChainToDomain(enterpriseCA, domain *graph.Node, groupExpansions impact.PathAggregator, cache ADCSCache, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], adcsEnabled bool) {

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostGoldenCert(ctx, tx, outC, domain, enterpriseCA); err != nil {
			log.Errorf("failed post processing for %s: %v", ad.GoldenCert.String(), err)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC1(ctx, tx, outC, groupExpansions, enterpriseCA, domain, cache); err != nil {
			log.Errorf("failed post processing for %s: %v", ad.ADCSESC1.String(), err)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC3(ctx, tx, outC, groupExpansions, enterpriseCA, domain, cache); err != nil {
			log.Errorf("failed post processing for %s: %v", ad.ADCSESC3.String(), err)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC6a(ctx, tx, outC, groupExpansions, enterpriseCA, domain, cache); err != nil {
			log.Errorf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
		}
		return nil
	})

	if adcsEnabled {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			if err := PostADCSESC6b(ctx, tx, outC, groupExpansions, enterpriseCA, domain, cache); err != nil {
				log.Errorf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
			}
			return nil
		})
	}

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC9a(ctx, tx, outC, groupExpansions, enterpriseCA, domain, cache); err != nil {
			log.Errorf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
		}
		return nil
	})

	if adcsEnabled {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			if err := PostADCSESC9b(ctx, tx, outC, groupExpansions, enterpriseCA, domain, cache); err != nil {
				log.Errorf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
			}
			return nil
		})
	}

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC10a(ctx, tx, outC, groupExpansions, enterpriseCA, domain, cache); err != nil {
			log.Errorf("failed post processing for %s: %v", ad.ADCSESC10a.String(), err)
		}
		return nil
	})

	if adcsEnabled {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			if err := PostADCSESC10b(ctx, tx, outC, groupExpansions, enterpriseCA, domain, cache); err != nil {
				log.Errorf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
			}
			return nil
		})
	}
}
