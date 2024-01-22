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

package neo4j

import (
	"sort"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
)

func updateKey(identityKind graph.Kind, identityProperties []string, updateKinds graph.Kinds) string {
	keys := []string{
		identityKind.String(),
	}

	keys = append(keys, identityProperties...)
	keys = append(keys, updateKinds.Strings()...)

	sort.Strings(keys)

	return strings.Join(keys, "")
}

func relUpdateKey(update graph.RelationshipUpdate) string {
	keys := []string{
		updateKey(update.Relationship.Kind, update.IdentityProperties, nil),
		updateKey(update.StartIdentityKind, update.StartIdentityProperties, update.Start.Kinds),
		updateKey(update.EndIdentityKind, update.EndIdentityProperties, update.End.Kinds),
	}

	sort.Strings(keys)

	return strings.Join(keys, "")
}

type relUpdates struct {
	identityKind            graph.Kind
	identityProperties      []string
	startIdentityKind       graph.Kind
	startIdentityProperties []string
	startNodeKindsToAdd     graph.Kinds
	endIdentityKind         graph.Kind
	endIdentityProperties   []string
	endNodeKindsToAdd       graph.Kinds
	properties              []map[string]any
}

type relUpdateByMap map[string]*relUpdates

func (s relUpdateByMap) add(update graph.RelationshipUpdate) {
	var (
		updateKey        = relUpdateKey(update)
		updateProperties = map[string]any{
			"r": update.Relationship.Properties.Map,
			"s": update.Start.Properties.Map,
			"e": update.End.Properties.Map,
		}
	)

	if updates, hasUpdates := s[updateKey]; hasUpdates {
		updates.properties = append(updates.properties, updateProperties)
	} else {
		s[updateKey] = &relUpdates{
			identityKind:            update.Relationship.Kind,
			identityProperties:      update.IdentityProperties,
			startIdentityKind:       update.StartIdentityKind,
			startIdentityProperties: update.StartIdentityProperties,
			startNodeKindsToAdd:     update.Start.Kinds,
			endIdentityKind:         update.EndIdentityKind,
			endIdentityProperties:   update.EndIdentityProperties,
			endNodeKindsToAdd:       update.End.Kinds,
			properties: []map[string]any{
				updateProperties,
			},
		}
	}
}

func cypherBuildRelationshipUpdateQueryBatch(updates []graph.RelationshipUpdate) ([]string, [][]map[string]any) {
	var (
		queries         []string
		queryParameters [][]map[string]any

		output         = strings.Builder{}
		batchedUpdates = relUpdateByMap{}
	)

	for _, update := range updates {
		batchedUpdates.add(update)
	}

	for _, batch := range batchedUpdates {
		output.WriteString("unwind $p as p merge (s:")
		output.WriteString(batch.startIdentityKind.String())

		if len(batch.startIdentityProperties) > 0 {
			output.WriteString(" {")

			firstIdentityProperty := true
			for _, identityProperty := range batch.startIdentityProperties {
				if firstIdentityProperty {
					firstIdentityProperty = false
				} else {
					output.WriteString(",")
				}

				output.WriteString(identityProperty)
				output.WriteString(":p.s.")
				output.WriteString(identityProperty)
			}

			output.WriteString("}")
		}

		output.WriteString(") merge (e:")
		output.WriteString(batch.endIdentityKind.String())

		if len(batch.endIdentityProperties) > 0 {
			output.WriteString(" {")

			firstIdentityProperty := true
			for _, identityProperty := range batch.endIdentityProperties {
				if firstIdentityProperty {
					firstIdentityProperty = false
				} else {
					output.WriteString(",")
				}

				output.WriteString(identityProperty)
				output.WriteString(":p.e.")
				output.WriteString(identityProperty)
			}

			output.WriteString("}")
		}

		output.WriteString(") merge (s)-[r:")
		output.WriteString(batch.identityKind.String())

		if len(batch.identityProperties) > 0 {
			output.WriteString(" {")

			firstIdentityProperty := true
			for _, identityProperty := range batch.identityProperties {
				if firstIdentityProperty {
					firstIdentityProperty = false
				} else {
					output.WriteString(",")
				}

				output.WriteString(identityProperty)
				output.WriteString(":p.r.")
				output.WriteString(identityProperty)
			}

			output.WriteString("}")
		}

		output.WriteString("]->(e) set s += p.s, e += p.e, r += p.r")

		if len(batch.startNodeKindsToAdd) > 0 {
			for _, kindToAdd := range batch.startNodeKindsToAdd {
				output.WriteString(", s:")
				output.WriteString(kindToAdd.String())
			}
		}

		if len(batch.endNodeKindsToAdd) > 0 {
			for _, kindToAdd := range batch.endNodeKindsToAdd {
				output.WriteString(", e:")
				output.WriteString(kindToAdd.String())
			}
		}

		output.WriteString(", s.lastseen = datetime({timezone: 'UTC'}), e.lastseen = datetime({timezone: 'UTC'});")

		// Write out the query to be run
		queries = append(queries, output.String())
		queryParameters = append(queryParameters, batch.properties)

		output.Reset()
	}

	return queries, queryParameters
}

type nodeUpdates struct {
	identityKind       graph.Kind
	identityProperties []string
	nodeKindsToAdd     graph.Kinds
	properties         []map[string]any
}

type nodeUpdateByMap map[string]*nodeUpdates

func (s nodeUpdateByMap) add(update graph.NodeUpdate) {
	updateKey := updateKey(update.IdentityKind, update.IdentityProperties, update.Node.Kinds)

	if updates, hasUpdates := s[updateKey]; hasUpdates {
		updates.properties = append(updates.properties, update.Node.Properties.Map)
	} else {
		s[updateKey] = &nodeUpdates{
			identityKind:       update.IdentityKind,
			identityProperties: update.IdentityProperties,
			nodeKindsToAdd:     update.Node.Kinds,
			properties: []map[string]any{
				update.Node.Properties.Map,
			},
		}
	}
}

func cypherBuildNodeUpdateQueryBatch(updates []graph.NodeUpdate) ([]string, []map[string]any) {
	var (
		queries         []string
		queryParameters []map[string]any

		output         = strings.Builder{}
		batchedUpdates = nodeUpdateByMap{}
	)

	for _, update := range updates {
		batchedUpdates.add(update)
	}

	for _, batch := range batchedUpdates {
		output.WriteString("unwind $p as p merge (n:")
		output.WriteString(batch.identityKind.String())

		if len(batch.identityProperties) > 0 {
			output.WriteString(" {")

			firstIdentityProperty := true
			for _, identityProperty := range batch.identityProperties {
				if firstIdentityProperty {
					firstIdentityProperty = false
				} else {
					output.WriteString(",")
				}

				output.WriteString(identityProperty)
				output.WriteString(":p.")
				output.WriteString(identityProperty)
			}

			output.WriteString("}")
		}

		output.WriteString(") set n += p")

		if len(batch.nodeKindsToAdd) > 0 {
			for _, kindToAdd := range batch.nodeKindsToAdd {
				output.WriteString(", n:")
				output.WriteString(kindToAdd.String())
			}
		}

		output.WriteString(";")

		// Write out the query to be run
		queries = append(queries, output.String())
		queryParameters = append(queryParameters, map[string]any{
			"p": batch.properties,
		})

		output.Reset()
	}

	return queries, queryParameters
}
