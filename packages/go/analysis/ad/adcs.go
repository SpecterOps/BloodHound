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

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
)

var ErrNoCertParent = errors.New("cert has no parent")

const (
	EkuAnyPurpose       = "2.5.29.37.0"
	EkuCertRequestAgent = "1.3.6.1.4.1.311.20.2.1"
)

func PostADCS(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (*post.AtomicPostProcessingStats, *ADCSCache, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing ADCS",
		attr.Namespace("analysis"),
		attr.Function("PostADCS"),
		attr.Scope("process"),
	)()

	var cache = NewADCSCache()

	if enterpriseCertAuthorities, err := FetchNodesByKind(ctx, db, ad.EnterpriseCA); err != nil {
		return &post.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed fetching enterpriseCA nodes: %w", err)
	} else if rootCertAuthorities, err := FetchNodesByKind(ctx, db, ad.RootCA); err != nil {
		return &post.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed fetching rootCA nodes: %w", err)
	} else if aiaCertAuthorities, err := FetchNodesByKind(ctx, db, ad.AIACA); err != nil {
		return &post.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed fetching AIACA nodes: %w", err)
	} else if certTemplates, err := FetchNodesByKind(ctx, db, ad.CertTemplate); err != nil {
		return &post.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed fetching cert template nodes: %w", err)
	} else if step1Stats, err := postADCSPreProcessStep1(ctx, db, enterpriseCertAuthorities, rootCertAuthorities, aiaCertAuthorities, certTemplates); err != nil {
		return &post.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed adcs pre-processing step 1: %w", err)
	} else if err := cache.BuildCache(ctx, db, enterpriseCertAuthorities, certTemplates); err != nil {
		return &post.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed building ADCS cache: %w", err)
	} else if step2Stats, err := postADCSPreProcessStep2(ctx, db, cache); err != nil {
		return &post.AtomicPostProcessingStats{}, cache, fmt.Errorf("failed adcs pre-processing step 2: %w", err)
	} else {
		defer measure.ContextMeasure(
			ctx,
			slog.LevelInfo,
			"ADCS ESC Processing",
			attr.Namespace("analysis"),
			attr.Function("PostADCS"),
			attr.Scope("routine"),
		)()

		operation := post.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing")

		operation.Stats.Merge(step1Stats)
		operation.Stats.Merge(step2Stats)

		for _, certChains := range cache.GetChainedDomains() {
			processEnterpriseCAWithValidCertChainToDomain(certChains, localGroupData, cache, operation)
		}

		return &operation.Stats, cache, operation.Done()
	}
}

// postADCSPreProcessStep1 processes the edges that are not dependent on any other post-processed edges
func postADCSPreProcessStep1(ctx context.Context, db graph.Database, enterpriseCertAuthorities, rootCertAuthorities, aiaCertAuthorities, certTemplates []*graph.Node) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"ADCS Post-processing Step 1",
		attr.Namespace("analysis"),
		attr.Function("postADCSPreProcessStep1"),
		attr.Scope("routine"),
	)()

	operation := post.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 1")

	if err := PostTrustedForNTAuth(ctx, db, operation); err != nil {
		operation.Done()
		return &post.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.TrustedForNTAuth.String(), err)
	} else if err := PostIssuedSignedBy(operation, enterpriseCertAuthorities, rootCertAuthorities, aiaCertAuthorities); err != nil {
		operation.Done()
		return &post.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.IssuedSignedBy.String(), err)
	} else if err := PostEnterpriseCAFor(operation, enterpriseCertAuthorities); err != nil {
		operation.Done()
		return &post.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnterpriseCAFor.String(), err)
	} else if err = PostExtendedByPolicyBinding(operation, certTemplates); err != nil {
		operation.Done()
		return &post.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.ExtendedByPolicy.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

// postADCSPreProcessStep2 Processes the edges that are dependent on those processed in postADCSPreProcessStep1
func postADCSPreProcessStep2(ctx context.Context, db graph.Database, cache *ADCSCache) (*post.AtomicPostProcessingStats, error) {
		defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"ADCS Post-processing Step 2",
		attr.Namespace("analysis"),
		attr.Function("postADCSPreProcessStep2"),
		attr.Scope("routine"),
	)()

	operation := post.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 2")

	if err := PostEnrollOnBehalfOf(cache, operation); err != nil {
		operation.Done()
		return &post.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnrollOnBehalfOf.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

func processEnterpriseCAWithValidCertChainToDomain(certChains *EnterpriseCAChainedDomains, localGroupData *LocalGroupData, cache *ADCSCache, operation post.StatTrackedOperation[post.EnsureRelationshipJob]) {

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing GoldenCert",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostGoldenCert(ctx, tx, outC, certChains); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for GoldenCert missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for GoldenCert",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC1",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC1(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC1 missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC1",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC3",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC3(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC3 missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC3",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC4",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC4(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC4 missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC4",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC6a",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC6a(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC6a missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC6a",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC6b",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC6b(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC6b missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC6b",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC9a",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC9a(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC9a missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC9a",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC9b",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC9b(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC9b missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC9b",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC10a",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC10a(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC10a missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC10a",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC10b",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC10b(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC10b missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC10b",
				attr.Error(err),
			)
		}
		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		defer measure.ContextMeasureWithThreshold(
			ctx,
			slog.LevelInfo,
			"Post-processing ADCSESC13",
			attr.Namespace("analysis"),
			attr.Function("processEnterpriseCAWithValidCertChainToDomain"),
			attr.Scope("routine"),
			slog.Uint64("enterprise_ca_id", uint64(certChains.EnterpriseCA.ID)),
		)()

		if err := PostADCSESC13(ctx, tx, outC, localGroupData, certChains, cache); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.WarnContext(
				ctx,
				"Post processing for ADCSESC13 missing property",
				attr.Error(err),
			)
		} else if err != nil {
			slog.ErrorContext(
				ctx,
				"Failed post processing for ADCSESC13",
				attr.Error(err),
			)
		}
		return nil
	})
}
