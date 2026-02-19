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
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
)

var (
	ErrNoCertParent     = errors.New("cert has no parent")
	EkuAnyPurpose       = "2.5.29.37.0"
	EkuCertRequestAgent = "1.3.6.1.4.1.311.20.2.1"
)

func PostADCS(ctx context.Context, db graph.Database, localGroupData *LocalGroupData, adcsEnabled bool) (*analysis.AtomicPostProcessingStats, ADCSCache, error) {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing ADCS",
		attr.Namespace("analysis"),
		attr.Function("PostADCS"),
		attr.Scope("process"),
	)()

	var cache = NewADCSCache()
	if enterpriseCertAuthorities, err := FetchNodesByKind(ctx, db, ad.EnterpriseCA); err != nil {
		return &analysis.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed fetching enterpriseCA nodes: %w", err)
	} else if rootCertAuthorities, err := FetchNodesByKind(ctx, db, ad.RootCA); err != nil {
		return &analysis.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed fetching rootCA nodes: %w", err)
	} else if aiaCertAuthorities, err := FetchNodesByKind(ctx, db, ad.AIACA); err != nil {
		return &analysis.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed fetching AIACA nodes: %w", err)
	} else if certTemplates, err := FetchNodesByKind(ctx, db, ad.CertTemplate); err != nil {
		return &analysis.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed fetching cert template nodes: %w", err)
	} else if step1Stats, err := postADCSPreProcessStep1(ctx, db, enterpriseCertAuthorities, rootCertAuthorities, aiaCertAuthorities, certTemplates); err != nil {
		return &analysis.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed adcs pre-processing step 1: %w", err)
	} else if err := cache.BuildCache(ctx, db, enterpriseCertAuthorities, certTemplates); err != nil {
		return &analysis.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed building ADCS cache: %w", err)
	} else if step2Stats, err := postADCSPreProcessStep2(ctx, db, cache); err != nil {
		return &analysis.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed adcs pre-processing step 2: %w", err)
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing")

		operation.Stats.Merge(step1Stats)
		operation.Stats.Merge(step2Stats)

		for _, enterpriseCA := range cache.GetEnterpriseCertAuthorities() {
			innerEnterpriseCA := enterpriseCA

			targetDomains := &graph.NodeSet{}
			for _, domain := range cache.GetDomains() {
				innerDomain := domain

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) && cache.DoesCAHaveHostingComputer(innerEnterpriseCA) {
					targetDomains.Add(innerDomain)
				}
			}
			processEnterpriseCAWithValidCertChainToDomain(innerEnterpriseCA, targetDomains, localGroupData, cache, operation)
		}
		return &operation.Stats, cache, operation.Done()
	}
}

// postADCSPreProcessStep1 processes the edges that are not dependent on any other post-processed edges
func postADCSPreProcessStep1(ctx context.Context, db graph.Database, enterpriseCertAuthorities, rootCertAuthorities, aiaCertAuthorities, certTemplates []*graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 1")
	// TODO clean up the operation.Done() calls below

	if err := PostTrustedForNTAuth(ctx, db, operation); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.TrustedForNTAuth.String(), err)
	} else if err := PostIssuedSignedBy(operation, enterpriseCertAuthorities, rootCertAuthorities, aiaCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.IssuedSignedBy.String(), err)
	} else if err := PostEnterpriseCAFor(operation, enterpriseCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnterpriseCAFor.String(), err)
	} else if err = PostExtendedByPolicyBinding(operation, certTemplates); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.ExtendedByPolicy.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

// postADCSPreProcessStep2 Processes the edges that are dependent on those processed in postADCSPreProcessStep1
func postADCSPreProcessStep2(ctx context.Context, db graph.Database, cache ADCSCache) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 2")

	if err := PostEnrollOnBehalfOf(cache, operation); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnrollOnBehalfOf.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

func processEnterpriseCAWithValidCertChainToDomain(enterpriseCA *graph.Node, targetDomains *graph.NodeSet, localGroupData *LocalGroupData, cache ADCSCache, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostGoldenCert(ctx, tx, outC, enterpriseCA, targetDomains); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.GoldenCert.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.GoldenCert.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC1(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC1.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC1.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC3(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC3.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC3.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC4(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC4.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC4.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC6a(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC6a.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC6a.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC6b(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC6b.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC6b.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC9a(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC9a.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC9a.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC9b(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC9b.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC9b.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC10a(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC10a.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC10a.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC10b(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC10b.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC10b.String(), err))
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if err := PostADCSESC13(ctx, tx, outC, localGroupData, enterpriseCA, targetDomains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(ctx, fmt.Sprintf("Post processing for %s: %v", ad.ADCSESC13.String(), err))
		} else if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed post processing for %s: %v", ad.ADCSESC13.String(), err))
		}
		return nil
	})
}
