// Copyright 2025 Specter Ops, Inc.
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

package graphify

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util"
)

func IngestNodes(ingestCtx *IngestContext, baseKind graph.Kind, nodes []ein.IngestibleNode) error {
	var (
		errs = util.NewErrorCollector()
	)

	for _, next := range nodes {
		if err := IngestNode(ingestCtx, baseKind, next); err != nil {
			slog.Error("Error ingesting node",
				slog.String("objectid", next.ObjectID),
				slog.String("err", err.Error()),
			)
			errs.Add(err)
		}
	}
	return errs.Combined()
}

func IngestNode(ic *IngestContext, baseKind graph.Kind, nextNode ein.IngestibleNode) error {
	var (
		nodeKinds            = MergeNodeKinds(baseKind, nextNode.Labels...)
		normalizedProperties = normalizeEinNodeProperties(nextNode.PropertyMap, nextNode.ObjectID, ic.IngestTime)
		nodeUpdate           = graph.NodeUpdate{
			Node:         graph.PrepareNode(graph.AsProperties(normalizedProperties), nodeKinds...),
			IdentityKind: baseKind,
			IdentityProperties: []string{
				common.ObjectID.String(),
			},
		}
	)

	if err := validateNodeKinds(nodeUpdate.Node.ID.String(), nodeKinds); err != nil {
		return err
	}

	return maybeSubmitNodeUpdate(ic, nodeUpdate)
}

// validateNodeKinds enforces basic invariants on a node's kinds.
//
// Rules:
//   - A node must have at least one kind.
//   - A node may not have more than 3 kinds.
func validateNodeKinds(objectID string, kinds graph.Kinds) error {
	switch {
	case len(kinds) == 0:
		slog.Warn("Skipping node with no kinds",
			slog.String("objectid", objectID),
			slog.Int("num_kinds", 0),
		)
		return fmt.Errorf("node %s has no kinds; at least 1 kind is required", objectID)

	case len(kinds) > 3:
		slog.Warn("Skipping node with too many kinds",
			slog.String("objectid", objectID),
			slog.Int("num_kinds", len(kinds)),
			slog.String("kinds", strings.Join(kinds.Strings(), ", ")),
		)
		return fmt.Errorf("node %s has too many kinds (%d); max allowed is 3", objectID, len(kinds))

	default:
		return nil
	}
}

// maybeSubmitNodeUpdate decides whether to upsert a node directly, or route it
// through the changelog for deduplication and caching.
func maybeSubmitNodeUpdate(ingestCtx *IngestContext, update graph.NodeUpdate) error {
	if !ingestCtx.HasChangelog() {
		// No changelog: always update via dawgs batch
		return ingestCtx.Batch.UpdateNodeBy(update)
	}

	objectid, err := update.Node.Properties.Get(common.ObjectID.String()).String()
	if err != nil {
		return fmt.Errorf("reading objectid failed: %w", err)
	}

	// Build change event for changelog
	change := changelog.NewNodeChange(
		objectid,
		update.Node.Kinds,
		update.Node.Properties,
	)

	shouldSubmit, err := ingestCtx.Manager.ResolveChange(change)
	if err != nil {
		return fmt.Errorf("resolve node change: %w", err)
	}

	if shouldSubmit {
		// New/modified: update via dawgs batch
		return ingestCtx.Batch.UpdateNodeBy(update)
	}

	// Unchanged: enqueue change-- this is needed to maintain reconciliation
	if ok := ingestCtx.Manager.Submit(ingestCtx.Ctx, change); !ok {
		slog.WarnContext(ingestCtx.Ctx, "Changelog submit dropped", slog.String("objectid", objectid))
	}

	return nil
}

func normalizeEinNodeProperties(properties map[string]any, objectID string, ingestTime time.Time) map[string]any {
	if properties == nil {
		properties = make(map[string]any)
	}
	delete(properties, ReconcileProperty)
	properties[common.LastSeen.String()] = ingestTime
	properties[common.ObjectID.String()] = strings.ToUpper(objectID)

	// Ensure that name, operatingsystem, and distinguishedname properties are upper case
	if rawName, hasName := properties[common.Name.String()]; hasName && rawName != nil {
		if name, typeMatches := rawName.(string); typeMatches {
			properties[common.Name.String()] = strings.ToUpper(name)
		} else {
			slog.Warn(
				"Bad type found for node name property during ingest",
				slog.String("objectid", objectID),
				slog.Group(
					"type_mismatch",
					slog.String("expected", "string"),
					slog.String("got", fmt.Sprintf("%T", rawName))))
		}
	}

	if rawOS, hasOS := properties[common.OperatingSystem.String()]; hasOS && rawOS != nil {
		if os, typeMatches := rawOS.(string); typeMatches {
			properties[common.OperatingSystem.String()] = strings.ToUpper(os)
		} else {
			slog.Warn(
				"Bad type found for node operating system property during ingest",
				slog.String("objectid", objectID),
				slog.Group(
					"type_mismatch",
					slog.String("expected", "string"),
					slog.String("got", fmt.Sprintf("%T", rawOS))))
		}
	}

	if rawDN, hasDN := properties[ad.DistinguishedName.String()]; hasDN && rawDN != nil {
		if dn, typeMatches := rawDN.(string); typeMatches {
			properties[ad.DistinguishedName.String()] = strings.ToUpper(dn)
		} else {
			slog.Warn(
				"Bad type found for node distinguished name property during ingest",
				slog.String("objectid", objectID),
				slog.Group(
					"type_mismatch",
					slog.String("expected", "string"),
					slog.String("got", fmt.Sprintf("%T", rawDN))))
		}
	}

	return properties
}
