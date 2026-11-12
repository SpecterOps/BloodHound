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

package hybrid

import (
	"context"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

// hostPair stores an unordered match with the smaller node ID first.
type hostPair struct {
	FirstID  graph.ID
	SecondID graph.ID
}

// hostIndex retains only node IDs under nonempty, case-normalized identifiers.
// Multiple nodes can share an identifier and must all participate in matching.
type hostIndex map[string][]graph.ID

func (s hostIndex) add(identifier string, nodeID graph.ID) {
	if identifier != "" {
		s[identifier] = append(s[identifier], nodeID)
	}
}

// lookup avoids allocating a normalized string for common ASCII identifiers.
// The temporary string conversion stays in the map lookup so the buffer can
// remain on the stack. Longer and non-ASCII identifiers retain hostIdentifier's
// normalization semantics; identifier syntax is not validated.
func (s hostIndex) lookup(value any) []graph.ID {
	var (
		identifier, isString = value.(string)
		buffer               [64]byte
		firstLower           = -1
	)
	if len(s) == 0 || !isString || identifier == "" {
		return nil
	}
	if len(identifier) > len(buffer) {
		return s[strings.ToUpper(identifier)]
	}
	for offset := range len(identifier) {
		var character = identifier[offset]
		if character >= 128 {
			return s[strings.ToUpper(identifier)]
		}
		if character >= 'a' && character <= 'z' {
			firstLower = offset
			break
		}
	}
	if firstLower == -1 {
		return s[identifier]
	}
	copy(buffer[:], identifier)
	for offset := firstLower; offset < len(identifier); offset++ {
		var character = identifier[offset]
		if character >= 128 {
			return s[strings.ToUpper(identifier)]
		}
		if character >= 'a' && character <= 'z' {
			buffer[offset] = character - ('a' - 'A')
		}
	}
	return s[string(buffer[:len(identifier)])]
}

// azureHostIndexes caches only Azure identifiers and node IDs. Projected rows
// are discarded after indexing; AD computers are matched one row at a time.
type azureHostIndexes struct {
	vmsByVMID           hostIndex
	vmsByResourceID     hostIndex
	devicesBySID        hostIndex
	devicesByResourceID hostIndex
}

func newAzureHostIndexes() azureHostIndexes {
	return azureHostIndexes{
		vmsByVMID:           make(hostIndex),
		vmsByResourceID:     make(hostIndex),
		devicesBySID:        make(hostIndex),
		devicesByResourceID: make(hostIndex),
	}
}

func (s azureHostIndexes) addVM(nodeID graph.ID, vmID, resourceID any) {
	// The Compute VM GUID (vm_id) and ARM resource ID (objectid) are independent identifiers.
	s.vmsByVMID.add(hostIdentifier(vmID), nodeID)
	s.vmsByResourceID.add(hostIdentifier(resourceID), nodeID)
}

func (s azureHostIndexes) addDevice(nodeID graph.ID, onPremID, physicalIDs any) {
	s.devicesBySID.add(hostIdentifier(onPremID), nodeID)
	for _, resourceID := range deviceVMResourceIDs(physicalIDs) {
		s.devicesByResourceID.add(resourceID, nodeID)
	}
}

func correlateHosts(ctx context.Context, tx graph.Transaction, sink *post.FilteredRelationshipSink) error {
	var (
		hosts, err = fetchAzureHostIndexes(tx)
		matches    = newHostPairSubmitter(func(relationship post.EnsureRelationshipJob) bool {
			return sink.Submit(ctx, relationship)
		})
	)
	if err != nil {
		return fmt.Errorf("fetching Azure hosts for correlation: %w", err)
	}

	// Azure resource references can match without any AD computers being present.
	if err := hosts.matchDevices(matches); err != nil {
		return err
	}
	// Only VM GUIDs and on-premises SIDs require an AD scan. Resource IDs do not.
	if len(hosts.vmsByVMID) == 0 && len(hosts.devicesBySID) == 0 {
		return nil
	}

	// Project only the required fields and submit matches immediately. No AD nodes
	// or computer indexes are retained, including computers with no Azure counterpart.
	if err := tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)).Query(func(results graph.Result) error {
		defer results.Close()
		for results.Next() {
			var (
				values          = results.Values()
				computerID, err = graph.AsNumeric[graph.ID](values[0])
			)
			if err != nil {
				return err
			}
			if err := hosts.matchComputer(computerID, values[1], values[2], matches); err != nil {
				return err
			}
		}
		return results.Error()
	}, query.Returning(query.NodeID(), query.NodeProperty(common.ObjectID.String()), query.NodeProperty("azvmid"))); err != nil {
		return fmt.Errorf("correlating AD computers: %w", err)
	}
	return nil
}

func fetchAzureHostIndexes(tx graph.Transaction) (azureHostIndexes, error) {
	var hosts = newAzureHostIndexes()
	if err := tx.Nodes().Filter(query.Kind(query.Node(), azure.VM)).Query(func(results graph.Result) error {
		defer results.Close()
		for results.Next() {
			var (
				values      = results.Values()
				nodeID, err = graph.AsNumeric[graph.ID](values[0])
			)
			if err != nil {
				return err
			}
			hosts.addVM(nodeID, values[1], values[2])
		}
		return results.Error()
	}, query.Returning(query.NodeID(), query.NodeProperty("vm_id"), query.NodeProperty(common.ObjectID.String()))); err != nil {
		return hosts, err
	}
	if err := tx.Nodes().Filter(query.Kind(query.Node(), azure.Device)).Query(func(results graph.Result) error {
		defer results.Close()
		for results.Next() {
			var (
				values      = results.Values()
				nodeID, err = graph.AsNumeric[graph.ID](values[0])
			)
			if err != nil {
				return err
			}
			hosts.addDevice(nodeID, values[1], values[2])
		}
		return results.Error()
	}, query.Returning(query.NodeID(), query.NodeProperty(azure.OnPremID.String()), query.NodeProperty("physical_ids"))); err != nil {
		return hosts, err
	}
	return hosts, nil
}

func (s azureHostIndexes) matchDevices(matches *hostPairSubmitter) error {
	for resourceID, deviceIDs := range s.devicesByResourceID {
		for _, deviceID := range deviceIDs {
			if err := matches.submitMatches(deviceID, s.vmsByResourceID[resourceID]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s azureHostIndexes) matchComputer(nodeID graph.ID, sid, vmID any, matches *hostPairSubmitter) error {
	// SharpHound's AzureRegistry collection provides azvmid on AD computers.
	if err := matches.submitMatches(nodeID, s.vmsByVMID.lookup(vmID)); err != nil {
		return err
	}
	return matches.submitMatches(nodeID, s.devicesBySID.lookup(sid))
}

type hostPairSubmitter struct {
	seen               map[hostPair]struct{}
	submitRelationship func(post.EnsureRelationshipJob) bool
}

func newHostPairSubmitter(submitRelationship func(post.EnsureRelationshipJob) bool) *hostPairSubmitter {
	return &hostPairSubmitter{seen: make(map[hostPair]struct{}), submitRelationship: submitRelationship}
}

func (s *hostPairSubmitter) submitMatches(sourceID graph.ID, targetIDs []graph.ID) error {
	for _, targetID := range targetIDs {
		if sourceID == targetID {
			continue
		}
		var pair = hostPair{FirstID: min(sourceID, targetID), SecondID: max(sourceID, targetID)}
		if _, seen := s.seen[pair]; seen {
			continue
		}
		// The DCA tracker does not deduplicate new edges within a run. Share this set
		// across every matching rule, and keep no additional slice of matched pairs.
		s.seen[pair] = struct{}{}
		for _, relationship := range [...]post.EnsureRelationshipJob{
			{FromID: pair.FirstID, ToID: pair.SecondID, Kind: common.SameHostAs},
			{FromID: pair.SecondID, ToID: pair.FirstID, Kind: common.SameHostAs},
		} {
			if !s.submitRelationship(relationship) {
				return fmt.Errorf("unable to submit %s relationship to the sink", common.SameHostAs)
			}
		}
	}
	return nil
}

// Missing, empty or non-string properties cannot match: indexes never contain
// empty keys, so lookups of "" return no targets. Syntax is intentionally not validated.
func hostIdentifier(value any) string {
	if identifier, isString := value.(string); isString {
		return strings.ToUpper(identifier)
	}
	return ""
}

func deviceVMResourceIDs(value any) []string {
	var (
		physicalIDs []string
		resourceIDs []string
	)
	// Both slice representations occur in graph drivers and collector payloads.
	// Ignore malformed entries individually so valid resource references survive.
	switch values := value.(type) {
	case []string:
		physicalIDs = values
	case []any:
		for _, value := range values {
			if identifier, isString := value.(string); isString {
				physicalIDs = append(physicalIDs, identifier)
			}
		}
	}
	for _, physicalID := range physicalIDs {
		var prefix, resourceID, found = strings.Cut(physicalID, ":")
		if found && strings.EqualFold(prefix, "[AzureResourceId]") && resourceID != "" {
			resourceIDs = append(resourceIDs, strings.ToUpper(resourceID))
		}
	}
	return resourceIDs
}
