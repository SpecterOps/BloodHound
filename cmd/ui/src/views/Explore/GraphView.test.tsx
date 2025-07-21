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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, waitFor } from 'src/test-utils';
import GraphView from './GraphView';
import { cypherResponse, featureFlagsResponse } from './graph-view-test-data';

const server = setupServer(
    rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(ctx.json(cypherResponse));
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(ctx.json(featureFlagsResponse));
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(ctx.status(200));
    }),
    rest.get('/api/v2/custom-nodes', (req, res, ctx) => {
        return res(ctx.status(200));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('GraphView', () => {
    it('renders a graph view', () => {
        render(<GraphView />, { route: `/graphview?searchType=cypher&cypherSearch=encodedquery` });
        const container = screen.getByTestId('explore');
        expect(container).toBeInTheDocument();
    });

    it('displays an error message', async () => {
        server.use(
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.status(500));
            })
        );
        console.error = vi.fn();
        render(<GraphView />);
        const errorAlert = await waitFor(() =>
            screen.findByText('An unexpected error has occurred. Please refresh the page and try again.')
        );

        expect(errorAlert).toBeInTheDocument();
    });

    it('renders a table if the query has NO node edges', async () => {
        const urlSearchParams = new URLSearchParams();
        urlSearchParams.append('searchType', 'cypher');
        urlSearchParams.append('cypherSearch', 'encodedquery');

        render(<GraphView />, { route: `/graphview?searchType=cypher&cypherSearch=encodedquery` });

        const table = await screen.findByRole('table');

        expect(table).toBeInTheDocument();

        const rows = await screen.findAllByRole('row');

        const expectedNumberOfRows = Object.keys(cypherResponse.data.nodes).length + 1; // plus one for the header row
        expect(rows.length).toBe(expectedNumberOfRows);
    });

    it('renders a graph if the query has any node edges', async () => {
        const urlSearchParams = new URLSearchParams();
        urlSearchParams.append('searchType', 'cypher');
        urlSearchParams.append('cypherSearch', 'encodedquery');

        const clonedCypherResponse = Object.assign({}, cypherResponse);

        clonedCypherResponse.data.edges.push({
            source: '108',
            target: '108',
            label: 'some label',
            kind: 'some kind',
            lastSeen: 'some lastSeen',
            impactPercent: 10,
            exploreGraphId: 'some exploreGraphId',
            data: {},
        });

        server.use(
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.json(clonedCypherResponse));
            })
        );

        render(<GraphView />, { route: `/graphview?searchType=cypher&cypherSearch=encodedquery` });

        await new Promise((res) => setTimeout(res, 2000));

        const sigma = await screen.findByTestId('sigma-container-wrapper');
        const table = screen.queryByRole('table');

        expect(sigma).toBeInTheDocument();
        expect(table).not.toBeInTheDocument();
    });
});
