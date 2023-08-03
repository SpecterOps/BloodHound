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
	"log"

	"github.com/specterops/bloodhound/lab"
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func NewDomainFixture(sid uuid.UUID, name string, collected, blocksInheritance bool) *lab.Fixture[*graph.Node] {
	fixture := lab.NewFixture(func(harness *lab.Harness) (*graph.Node, error) {
		if node, err := CreateNode(harness, graph.PropertyMap{
			common.Name:          name,
			common.ObjectID:      sid.String(),
			ad.DomainSID:         sid.String(),
			common.Collected:     collected,
			ad.BlocksInheritance: blocksInheritance,
		}, ad.Entity, ad.Domain); err != nil {
			return nil, err
		} else {
			return node, nil
		}
	}, func(harness *lab.Harness, node *graph.Node) error {
		if err := DeleteNode(harness, node); err != nil {
			return err
		} else {
			return nil
		}
	})

	if err := lab.SetDependency(fixture, GraphDBFixture); err != nil {
		log.Fatalln(err)
	}

	return fixture
}

var (
	BasicDomainSID     = uuid.Must(uuid.NewV4())
	BasicDomainName    = "TestDomain"
	BasicDomainFixture = NewDomainFixture(BasicDomainSID, BasicDomainName, true, false)
)
