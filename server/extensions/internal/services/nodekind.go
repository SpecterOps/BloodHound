// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"fmt"
	"time"
)

// ErrNodeKindNotFound indicates that no node kind exists for the requested id.
var ErrNodeKindNotFound = errors.New("node kind not found")

// NodeKind is the domain representation of a node kind, combining the schema_node_kinds
// row with the kind name resolved from the kind table. Info is populated by the service.
type NodeKind struct {
	ID                int32
	SchemaExtensionID int32
	KindID            int32
	Name              string // from kind.name
	DisplayName       string
	Description       string
	IsDisplayKind     bool
	Icon              string
	Color             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	Info              []KindInfo
	Extension         Extension
}

// GetNodeKind returns the node kind identified by id with its entity-panel infos and
// owning extension attached. The extension is optional: when no matching extension exists
// the node kind is returned with a zero-value Extension. ErrNodeKindNotFound bubbles up
// from the store when no node kind matches.
func (s *Service) GetNodeKind(ctx context.Context, id int32) (NodeKind, error) {

	if nodeKind, err := s.db.GetNodeKind(ctx, id); err != nil {
		return NodeKind{}, err
	} else if nodeKind.Info, err = s.db.GetKindInfos(ctx, nodeKind.Name); err != nil {
		return NodeKind{}, fmt.Errorf("fetching kind infos for node kind %s: %w", nodeKind.Name, err)
	} else {
		if nodeKind.SchemaExtensionID != 0 {
			if nodeKind.Extension, err = s.db.GetExtension(ctx, nodeKind.SchemaExtensionID); err != nil && !errors.Is(err, ErrExtensionNotFound) {
				return NodeKind{}, fmt.Errorf("fetching extension %d for node kind %d: %w", nodeKind.SchemaExtensionID, id, err)
			}
		}
		return nodeKind, nil
	}
}
