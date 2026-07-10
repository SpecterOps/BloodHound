// Copyright 2026 Specter Ops, Inc.
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

package services

import (
	"context"
	"errors"
	"slices"

	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/slicesext"
)

// Node is the domain representation of a graph node together with
// its resolved kinds and properties.
type Node struct {
	ID         int64
	Kinds      []Kind
	Properties map[string]any
}

// KindNames returns the list of kinds in string form
func (s Node) KindNames() []string {
	return slicesext.Map(s.Kinds, func(kind Kind) string {
		return kind.Name
	})
}

// EnvironmentID returns the environment id for the node if set.
// For Azure nodes, this returns a tenant ID.
// For Active Directory nodes, this returns a domain SID.
// For all other types of nodes, this returns the "environmentid" property.
// For all nodes, the returned boolean will be false if the environment property is not set or improperly declared.
func (s Node) EnvironmentID() (string, bool) {
	var (
		kinds = s.KindNames()
		key   = graphschema.EnvironmentIDKey
	)

	if slices.Contains(kinds, ad.Entity.String()) {
		key = ad.DomainSID.String()
	} else if slices.Contains(kinds, azure.Entity.String()) {
		key = azure.TenantID.String()
	}

	id, ok := s.Properties[key].(string)
	return id, ok
}

// ErrNodeNotFound indicates that no graph node exists for the requested id.
var ErrNodeNotFound = errors.New("node not found")

// GetNode returns the node identified by the graph-assigned id with its
// kinds resolved to the integer identifiers from the schema_node_kinds table.
// ErrNodeNotFound is returned when the node does not exist. Kinds that are not
// registered in schema_node_kinds are returned with ID=nil (best-effort resolution).
func (s *Service) GetNode(ctx context.Context, id int64) (Node, error) {
	var (
		node      Node
		kindNames []string
		kinds     []Kind
		err       error
	)

	if node, err = s.db.GetNode(ctx, id); err != nil {
		return Node{}, err
	}

	for _, kind := range node.Kinds {
		kindNames = append(kindNames, kind.Name)
	}

	if kinds, err = s.db.GetNodeKindsByNames(ctx, kindNames); err != nil {
		return Node{}, err
	} else {
		node.Kinds = kinds
		return node, nil
	}
}
