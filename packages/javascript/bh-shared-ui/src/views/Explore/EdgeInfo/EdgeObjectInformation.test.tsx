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
import { NodeDetails, RelationshipDetails } from 'js-client-library';
import { render, screen } from '../../../test-utils';
import { ObjectInfoPanelContextProvider } from '../providers';
import EdgeObjectInformation from './EdgeObjectInformation';

const selectedEdge: RelationshipDetails = {
    relationship_id: 1,
    kind: { name: 'DCSync', relationship_kind_id: 1 },
    properties: { lastSeen: '2023-09-07T11:10:33.664596893Z', is_traversable: true, isacl: false },
    source_node_id: 1,
    target_node_id: 2,
};

const mockSourceNode: NodeDetails = {
    node_id: 1,
    kinds: [],
    properties: { objectid: 'source-objectid', name: 'source_node', lastSeen: '2023-09-07T11:10:33.664596893Z' },
};

const mockTargetNode: NodeDetails = {
    node_id: 2,
    kinds: [],
    properties: { objectid: 'target-objectid', name: 'target_node', lastSeen: '2023-09-07T11:10:33.664596893Z' },
};

const EdgeObjectInformationWithProvider = () => (
    <ObjectInfoPanelContextProvider>
        <EdgeObjectInformation selectedEdge={selectedEdge} sourceNode={mockSourceNode} targetNode={mockTargetNode} />
    </ObjectInfoPanelContextProvider>
);

describe('EdgeObjectInformation', () => {
    test('The edge properties are fetched through the relationship endpoint which includes lastseen', async () => {
        render(<EdgeObjectInformationWithProvider />);

        expect(await screen.findByText(/source_node/)).toBeInTheDocument();

        //Display the information obtained from the relationship endpoint for all edge properties
        expect(screen.getByText(/Source Node:/)).toBeInTheDocument();
        expect(screen.getByText(/source_node/)).toBeInTheDocument();
        expect(screen.getByText(/Target Node:/)).toBeInTheDocument();
        expect(screen.getByText(/target_node/)).toBeInTheDocument();
        expect(screen.getByText(/Is ACL:/)).toBeInTheDocument();
        expect(screen.getByText(/FALSE/)).toBeInTheDocument();
        expect(screen.getByText(/Last Seen by BloodHound:/)).toBeInTheDocument();
    });
});
