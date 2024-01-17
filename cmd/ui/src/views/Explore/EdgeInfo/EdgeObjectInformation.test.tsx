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

import { SelectedEdge } from 'bh-shared-ui';
import EdgeObjectInformation from 'src/views/Explore/EdgeInfo/EdgeObjectInformation';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from 'src/test-utils';

const server = setupServer();

const selectedEdge: SelectedEdge = {
    id: '1',
    name: 'DCSync',
    data: { lastseen: '2023-09-07T11:10:33.664596893Z' },
    sourceNode: {
        name: 'source_node',
        id: '1',
        objectId: '1',
        type: 'User',
    },
    targetNode: { name: 'target_node', id: '2', objectId: '2', type: 'User' },
};

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EdgeObjectInformation', () => {
    test('The edge properties are fetched through a cypher search which includes lastseen', async () => {
        server.use(
            rest.post(`/api/v2/graphs/cypher`, (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            nodes: {},
                            edges: [
                                {
                                    source: '48297',
                                    target: '1238',
                                    label: 'DCSync',
                                    kind: 'DCSync',
                                    lastSeen: '2023-09-07T11:10:33.664596893Z',
                                    properties: {
                                        lastseen: '2023-09-07T11:10:33.664596893Z',
                                        isacl: false,
                                    },
                                },
                            ],
                        },
                    })
                );
            })
        );

        render(<EdgeObjectInformation selectedEdge={selectedEdge} />);

        expect(await screen.findByText(/source_node/)).toBeInTheDocument();

        //Display the information obtained from the cypher query for all edge properties
        expect(screen.getByText(/Source Node:/)).toBeInTheDocument();
        expect(screen.getByText(/source_node/)).toBeInTheDocument();
        expect(screen.getByText(/Target Node:/)).toBeInTheDocument();
        expect(screen.getByText(/target_node/)).toBeInTheDocument();
        expect(screen.getByText(/Is ACL:/)).toBeInTheDocument();
        expect(screen.getByText(/FALSE/)).toBeInTheDocument();
        expect(screen.getByText(/Last Collected by BloodHound:/)).toBeInTheDocument();
    });

    test('Error handling for fetching edge information', async () => {
        console.error = vi.fn();
        server.use(
            rest.post(`/api/v2/graphs/cypher`, (req, res, ctx) => {
                return res(
                    ctx.status(500),
                    ctx.json({
                        errorMessage: `Internal Server Error`,
                    })
                );
            })
        );

        render(<EdgeObjectInformation selectedEdge={selectedEdge} />);

        expect(await screen.findByText(/source_node/)).toBeInTheDocument();

        //These fields are all obtainable from the original graph response information
        //so if there is an error we can still display at least this information
        expect(screen.getByText(/Source Node:/)).toBeInTheDocument();
        expect(screen.getByText(/source_node/)).toBeInTheDocument();
        expect(screen.getByText(/Target Node:/)).toBeInTheDocument();
        expect(screen.getByText(/target_node/)).toBeInTheDocument();

        //These are extra fields that don't come with the graph response
        //so if there is an error with the edge query they will not be displayed
        expect(screen.queryByText(/Is ACL:/)).toBeNull();
        expect(screen.queryByText(/FALSE/)).toBeNull();
    });
});
