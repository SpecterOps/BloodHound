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

package nodeprops

import (
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

// ReadDomainIDandNameAsString extracts the domain SID and domain name from the given node. This function is
// intentionally placed in the internal package to avoid import cycle issues that would arise if it were implemented as
// an unexported function within the ad package and subsequently referenced by ad_test.go as of commit '87afb00a'. Using
// the internal package pattern here preserves proper package boundaries while still allowing ad_test.go to access the
// functionality needed for data generation and testing.
func ReadDomainIDandNameAsString(nodeToRead *graph.Node) (string, string, error) {
	if nodeToRead == nil {
		return "", "", fmt.Errorf("given nodeToRead is nil")
	}

	domainSID := nodeToRead.Properties.Get(ad.DomainSID.String())
	if domainSID.IsNil() {
		return "", "", fmt.Errorf("read domain SID property value is nil")
	}

	domainSIDStr, err := domainSID.String()
	if err != nil {
		return "", "", fmt.Errorf("failed to convert domainSID to string: %s", err)
	}

	if len(strings.TrimSpace(domainSIDStr)) == 0 {
		return "", "", fmt.Errorf("read domain SID is empty or blank")
	}

	domainName := nodeToRead.Properties.Get(common.Name.String())
	if domainName.IsNil() {
		return "", "", fmt.Errorf("read domain name property value is nil")
	}

	domainNameStr, err := domainName.String()
	if err != nil {
		return "", "", fmt.Errorf("failed to convert domain name to string: %s", err)
	}

	if len(strings.TrimSpace(domainNameStr)) == 0 {
		return "", "", fmt.Errorf("read domain name is empty or blank")
	}

	return domainSIDStr, domainNameStr, nil
}
