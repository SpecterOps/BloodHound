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
	"fmt"
	"log"
	"time"

	"github.com/specterops/bloodhound/lab"
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func NewComputerFixture(
	objectId uuid.UUID,
	name string,
	domainFixture *lab.Fixture[*graph.Node]) *lab.Fixture[*graph.Node] {

	fixture := NewGraphNodeFixture(func(h *lab.Harness) (graph.PropertyMap, error) {
		if domain, ok := lab.Unpack(h, domainFixture); !ok {
			return nil, fmt.Errorf("unable to unpack domain fixture")
		} else if domainSid, err := domain.Properties.Get(common.ObjectID.String()).String(); err != nil {
			return nil, fmt.Errorf("unable to unpack domain SID from fixture: %w", err)
		} else {
			return graph.PropertyMap{
				common.Name:            name,
				common.ObjectID:        objectId.String(),
				ad.DomainSID:           domainSid,
				common.OperatingSystem: "fake os",
				common.Enabled:         true,
				common.PasswordLastSet: time.Now().Unix(),
				common.LastSeen:        time.Now().UTC(),
			}, nil
		}
	}, ad.Entity, ad.Computer)

	if err := lab.SetDependency(fixture, domainFixture); err != nil {
		log.Fatalln(err)
	}
	return fixture
}

var (
	BasicComputerSID     = uuid.Must(uuid.NewV4())
	BasicComputerName    = "TestComputer"
	BasicComputerFixture = NewComputerFixture(BasicComputerSID, BasicComputerName, BasicDomainFixture)
)
