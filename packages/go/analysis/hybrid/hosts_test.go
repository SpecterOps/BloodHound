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
	"fmt"
	"strings"
	"testing"

	graph_mocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	hostTestVMID       = "abcdef01-2345-6789-abcd-0123456789ab"
	hostTestSID        = "S-1-5-21-111-222-333-1001"
	hostTestResourceID = "/subscriptions/abcdef01-2345-6789-abcd-0123456789ab/resourceGroups/Production/providers/Microsoft.Compute/virtualMachines/AADJoined"
)

func hostTestNode(id graph.ID, kind graph.Kind, properties map[string]any) *graph.Node {
	return graph.NewNode(id, graph.AsProperties(properties), kind)
}

func TestMatchingHostPairs(t *testing.T) {
	var (
		computer = hostTestNode(1, ad.Computer, map[string]any{"objectid": hostTestSID, "azvmid": strings.ToUpper(hostTestVMID)})
		vm       = hostTestNode(2, azure.VM, map[string]any{"objectid": strings.ToUpper(hostTestResourceID), "vm_id": hostTestVMID})
		device   = hostTestNode(3, azure.Device, map[string]any{
			"onpremid": strings.ToLower(hostTestSID),
			"physical_ids": []any{
				"[Other]:" + hostTestResourceID,
				"[AzureResourceId]:",
				123,
				"[AzureResourceId]:malformed",
				"[AzureResourceId]:" + hostTestResourceID,
				"[azureresourceid]:" + strings.ToUpper(hostTestResourceID),
			},
		})
	)

	for _, testCase := range []struct {
		name     string
		nodes    []*graph.Node
		expected []hostPair
	}{
		{name: "computer and VM without device", nodes: []*graph.Node{computer, vm}, expected: []hostPair{{FirstID: 1, SecondID: 2}}},
		{name: "VM and device without computer", nodes: []*graph.Node{vm, device}, expected: []hostPair{{FirstID: 2, SecondID: 3}}},
		{name: "device and computer without VM", nodes: []*graph.Node{device, computer}, expected: []hostPair{{FirstID: 1, SecondID: 3}}},
		{name: "all rules and duplicate resource IDs", nodes: []*graph.Node{device, vm, computer}, expected: []hostPair{{FirstID: 1, SecondID: 2}, {FirstID: 1, SecondID: 3}, {FirstID: 2, SecondID: 3}}},
		{name: "computer only", nodes: []*graph.Node{computer}},
		{name: "VM only", nodes: []*graph.Node{vm}},
		{name: "device only", nodes: []*graph.Node{device}},
		{name: "empty graph"},
		{
			name:     "all matching nodes are retained",
			nodes:    []*graph.Node{computer, vm, hostTestNode(4, ad.Computer, map[string]any{"azvmid": hostTestVMID})},
			expected: []hostPair{{FirstID: 1, SecondID: 2}, {FirstID: 2, SecondID: 4}},
		},
		{
			name:  "identifiers do not match other kinds",
			nodes: []*graph.Node{computer, hostTestNode(4, azure.User, map[string]any{"onpremid": hostTestSID, "vm_id": hostTestVMID})},
		},
		{
			name:     "no transitive inference",
			nodes:    []*graph.Node{computer, vm, hostTestNode(4, azure.Device, map[string]any{"onpremid": hostTestSID})},
			expected: []hostPair{{FirstID: 1, SecondID: 2}, {FirstID: 1, SecondID: 4}},
		},
		{
			name: "SID correlation survives malformed physical IDs",
			nodes: []*graph.Node{computer, hostTestNode(4, azure.Device, map[string]any{
				"onpremid": hostTestSID, "physical_ids": false,
			})},
			expected: []hostPair{{FirstID: 1, SecondID: 4}},
		},
		{
			name: "resource correlation survives malformed SID",
			nodes: []*graph.Node{vm, hostTestNode(4, azure.Device, map[string]any{
				"onpremid": 123, "physical_ids": []string{"[AzureResourceId]:" + hostTestResourceID},
			})},
			expected: []hostPair{{FirstID: 2, SecondID: 4}},
		},
		{
			name: "device matches every referenced VM",
			nodes: []*graph.Node{vm,
				hostTestNode(4, azure.VM, map[string]any{"objectid": hostTestResourceID + "2"}),
				hostTestNode(5, azure.Device, map[string]any{"physical_ids": []string{
					"[AzureResourceId]:" + hostTestResourceID,
					"[Other]:" + hostTestResourceID,
					"[AzureResourceId]:" + hostTestResourceID + "2",
				}}),
			},
			expected: []hostPair{{FirstID: 2, SecondID: 5}, {FirstID: 4, SecondID: 5}},
		},
		{
			name: "well-formed unmatched identifiers",
			nodes: []*graph.Node{computer,
				hostTestNode(4, azure.VM, map[string]any{"vm_id": "11111111-2222-3333-4444-555555555555"}),
				hostTestNode(5, azure.Device, map[string]any{
					"onpremid": hostTestSID + "0", "physical_ids": []string{"[AzureResourceId]:" + hostTestResourceID + "2"},
				}),
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var actual = collectMatchingHostPairs(t, testCase.nodes)
			require.ElementsMatch(t, testCase.expected, actual)
		})
	}
}

func TestMatchingHostPairsSkipsInvalidProperties(t *testing.T) {
	for _, value := range []any{nil, "", 123, true, []string{"wrong type"}} {
		t.Run(fmt.Sprintf("%T_%v", value, value), func(t *testing.T) {
			var nodes = []*graph.Node{
				hostTestNode(1, ad.Computer, map[string]any{"azvmid": value, "objectid": value}),
				hostTestNode(2, azure.VM, map[string]any{"vm_id": value}),
				hostTestNode(3, azure.Device, map[string]any{"onpremid": value, "physical_ids": value}),
			}
			require.Empty(t, collectMatchingHostPairs(t, nodes))
		})
	}
	require.Empty(t, collectMatchingHostPairs(t, []*graph.Node{
		hostTestNode(1, ad.Computer, map[string]any{}),
		hostTestNode(2, azure.VM, map[string]any{}),
		hostTestNode(3, azure.Device, map[string]any{}),
	}))
}

func TestMatchingHostPairsDoesNotValidateIdentifierSyntax(t *testing.T) {
	var nodes = []*graph.Node{
		hostTestNode(1, ad.Computer, map[string]any{"azvmid": "not-a-guid", "objectid": "not-a-sid"}),
		hostTestNode(2, azure.VM, map[string]any{"vm_id": "NOT-A-GUID", "objectid": "not-an-arm-resource-id"}),
		hostTestNode(3, azure.Device, map[string]any{
			"onpremid":     "NOT-A-SID",
			"physical_ids": []any{nil, 123, "[Other]:not-an-arm-resource-id", "[AzureResourceId]:", "[azureresourceid]:NOT-AN-ARM-RESOURCE-ID"},
		}),
	}

	require.ElementsMatch(t, []hostPair{
		{FirstID: 1, SecondID: 2},
		{FirstID: 1, SecondID: 3},
		{FirstID: 2, SecondID: 3},
	}, collectMatchingHostPairs(t, nodes))
}

func TestHostMatchingSkipsNonArrayPhysicalIDs(t *testing.T) {
	for _, value := range []any{nil, "", "[AzureResourceId]:" + hostTestResourceID, true, 123, map[string]any{"resource": hostTestResourceID}} {
		t.Run(fmt.Sprintf("%T_%v", value, value), func(t *testing.T) {
			require.Empty(t, collectMatchingHostPairs(t, []*graph.Node{
				hostTestNode(1, azure.VM, map[string]any{"objectid": hostTestResourceID}),
				hostTestNode(2, azure.Device, map[string]any{"physical_ids": value}),
			}))
		})
	}
}

func TestDeviceVMResourceIDs(t *testing.T) {
	for _, entries := range []any{
		[]string{"[Other]:" + hostTestResourceID, "[AzureResourceId]:" + hostTestResourceID},
		[]any{nil, false, "[Other]:" + hostTestResourceID, "[AzureResourceId]:" + hostTestResourceID},
	} {
		require.Equal(t, []string{strings.ToUpper(hostTestResourceID)}, deviceVMResourceIDs(
			entries,
		))
	}
}

// Exercise the same compact indexes and submission path without a database.
func collectMatchingHostPairs(t *testing.T, nodes []*graph.Node) []hostPair {
	t.Helper()
	var (
		hosts             = newAzureHostIndexes()
		pairs             []hostPair
		relationshipCount int
		matches           = newHostPairSubmitter(func(relationship post.EnsureRelationshipJob) bool {
			relationshipCount++
			if relationship.FromID < relationship.ToID {
				pairs = append(pairs, hostPair{FirstID: relationship.FromID, SecondID: relationship.ToID})
			}
			return true
		})
	)
	for _, node := range nodes {
		if node.Kinds.ContainsOneOf(azure.VM) {
			hosts.addVM(node.ID, node.Properties.Map["vm_id"], node.Properties.Map["objectid"])
		}
		if node.Kinds.ContainsOneOf(azure.Device) {
			hosts.addDevice(node.ID, node.Properties.Map["onpremid"], node.Properties.Map["physical_ids"])
		}
	}
	require.NoError(t, hosts.matchDevices(matches))
	for _, node := range nodes {
		if node.Kinds.ContainsOneOf(ad.Computer) {
			require.NoError(t, hosts.matchComputer(node.ID, node.Properties.Map["objectid"], node.Properties.Map["azvmid"], matches))
		}
	}
	require.Equal(t, 2*len(pairs), relationshipCount)
	return pairs
}

func TestCorrelateHostsScansADOnlyForVMIDsOrSIDs(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		vms     [][]any
		devices [][]any
		scanAD  bool
	}{
		{name: "no Azure hosts"},
		{name: "empty identifiers", vms: [][]any{{int64(1), "", ""}}, devices: [][]any{{int64(2), "", []string{}}}},
		{name: "incorrectly typed identifiers", vms: [][]any{{int64(1), false, 123}}, devices: [][]any{{int64(2), 123, false}}},
		{name: "VM ARM resource ID only", vms: [][]any{{int64(1), nil, hostTestResourceID}}},
		{name: "device resource reference only", devices: [][]any{{int64(2), nil, []string{"[AzureResourceId]:" + hostTestResourceID}}}},
		{name: "VM GUID", vms: [][]any{{int64(1), hostTestVMID, nil}}, scanAD: true},
		{name: "device SID", devices: [][]any{{int64(2), hostTestSID, nil}}, scanAD: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				controller  = gomock.NewController(t)
				transaction = graph_mocks.NewMockTransaction(controller)
				scans       = []struct {
					kind graph.Kind
					rows [][]any
				}{{azure.VM, testCase.vms}, {azure.Device, testCase.devices}}
			)
			if testCase.scanAD {
				scans = append(scans, struct {
					kind graph.Kind
					rows [][]any
				}{kind: ad.Computer})
			}
			for _, scan := range scans {
				var (
					nodeQuery = graph_mocks.NewMockNodeQuery(controller)
					result    = &hostProjectionResult{rows: scan.rows}
				)
				transaction.EXPECT().Nodes().Return(nodeQuery)
				nodeQuery.EXPECT().Filter(query.Kind(query.Node(), scan.kind)).Return(nodeQuery)
				nodeQuery.EXPECT().Query(gomock.Any(), gomock.Any()).DoAndReturn(func(delegate func(graph.Result) error, _ ...graph.Criteria) error {
					return delegate(result)
				})
			}
			// None of these fixtures can produce an edge, so no sink is needed. An
			// unexpected AD scan or edge submission fails the test.
			require.NoError(t, correlateHosts(t.Context(), transaction, nil))
		})
	}
}

type hostProjectionResult struct {
	graph.Result
	rows [][]any
	next int
}

func (s *hostProjectionResult) Next() bool {
	if s.next == len(s.rows) {
		return false
	}
	s.next++
	return true
}

func (s *hostProjectionResult) Values() []any { return s.rows[s.next-1] }
func (s *hostProjectionResult) Error() error  { return nil }
func (s *hostProjectionResult) Close()        {}

func TestHostPairSubmitterDeduplicatesWithinRun(t *testing.T) {
	var (
		relationships []post.EnsureRelationshipJob
		matches       = newHostPairSubmitter(func(relationship post.EnsureRelationshipJob) bool {
			relationships = append(relationships, relationship)
			return true
		})
	)
	require.NoError(t, matches.submitMatches(2, []graph.ID{2, 1, 1}))
	require.NoError(t, matches.submitMatches(1, []graph.ID{2}))
	require.Equal(t, []post.EnsureRelationshipJob{
		{FromID: 1, ToID: 2, Kind: common.SameHost},
		{FromID: 2, ToID: 1, Kind: common.SameHost},
	}, relationships)
}

func TestHostPairSubmitterPropagatesSinkFailure(t *testing.T) {
	var matches = newHostPairSubmitter(func(post.EnsureRelationshipJob) bool { return false })
	require.ErrorContains(t, matches.submitMatches(1, []graph.ID{2}), "unable to submit SameHost relationship")
}

func TestHostMatchingDeduplicatesAcrossRules(t *testing.T) {
	var nodes = []*graph.Node{
		graph.NewNode(1, graph.AsProperties(map[string]any{
			"objectid": hostTestSID, "azvmid": hostTestVMID, "vm_id": hostTestVMID,
		}), ad.Computer, azure.VM),
		graph.NewNode(2, graph.AsProperties(map[string]any{
			"onpremid": hostTestSID, "vm_id": hostTestVMID,
			"physical_ids": []string{"[AzureResourceId]:" + hostTestSID},
		}), azure.Device, azure.VM),
	}
	require.Equal(t, []hostPair{{FirstID: 1, SecondID: 2}}, collectMatchingHostPairs(t, nodes))
}
