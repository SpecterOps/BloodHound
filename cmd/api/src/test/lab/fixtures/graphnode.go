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

package fixtures

import (
	"context"
	"fmt"
	"log"

	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
)

type propsConstraint interface {
	graph.PropertyMap | map[graph.String]any | map[string]any
}

func NewGraphNodeFixture[Props propsConstraint](setup func(*lab.Harness) (Props, error), kinds ...graph.Kind) *lab.Fixture[*graph.Node] {
	fixture := lab.NewFixture(func(harness *lab.Harness) (*graph.Node, error) {
		if props, err := setup(harness); err != nil {
			return nil, err
		} else {
			return CreateNode(harness, props, kinds...)
		}
	}, func(harness *lab.Harness, node *graph.Node) error {
		return DeleteNode(harness, node)
	})
	if err := lab.SetDependency(fixture, GraphDBFixture); err != nil {
		log.Fatalln(err)
	}
	return fixture
}

func CreateNode[Props propsConstraint](harness *lab.Harness, props Props, kinds ...graph.Kind) (*graph.Node, error) {
	if graphdb, ok := lab.Unpack(harness, GraphDBFixture); !ok {
		return nil, fmt.Errorf("unable to unpack GraphDBFixture")
	} else {
		var out *graph.Node
		err := graphdb.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
			node, err := tx.CreateNode(graph.AsProperties(props), kinds...)
			out = node
			return err
		})
		return out, err
	}
}

func DeleteNode(harness *lab.Harness, node *graph.Node) error {
	if graphdb, ok := lab.Unpack(harness, GraphDBFixture); !ok {
		return fmt.Errorf("unable to unpack GraphDBFixture")
	} else {
		return graphdb.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
			return tx.Nodes().Filter(query.Equals(query.NodeID(), node.ID)).Delete()
		})
	}
}
