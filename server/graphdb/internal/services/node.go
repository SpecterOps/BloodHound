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
)

// Node is the domain representation of a graph node together with
// its resolved kinds and properties.
type Node struct {
	ID         int64
	Kinds      []Kind
	Properties map[string]any
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
