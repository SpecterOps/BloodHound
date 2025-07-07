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

import userEvent from '@testing-library/user-event';
import { Permission, createAuthStateWithPermissions } from 'bh-shared-ui';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, waitFor } from 'src/test-utils';
import ContextMenu from './ContextMenuZoneManagementEnabled';

const server = setupServer(
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(
            ctx.json({
                data: createAuthStateWithPermissions([Permission.GRAPH_DB_WRITE]).user,
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    tags: [
                        {
                            id: 2,
                            type: 3,
                            kind_id: 2,
                            name: 'Owned',
                            description: 'Owned',
                        },
                        {
                            id: 1,
                            type: 1,
                            kind_id: 1,
                            position: 1,
                            name: 'High Value',
                            description: 'High Value',
                        },
                    ],
                },
            })
        );
    }),
    rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(ctx.json({ data: { nodes: { abc: { objectId: 'abc' }, def: { objectId: 'def' } }, edges: [] } }));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('ContextMenu', () => {
    it('renders asset group edit options with graph write permissions', async () => {
        render(<ContextMenu contextMenu={{ mouseX: 0, mouseY: 0 }} onClose={vi.fn()} />, {
            route: '/test?selectedItem=abc',
        });

        const startNodeOption = await waitFor(() => screen.findByRole('menuitem', { name: /set as starting node/i }));
        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        const addToHighValue = screen.getByRole('menuitem', { name: /add to high value/i });
        const addToOwned = screen.getByRole('menuitem', { name: /add to owned/i });

        expect(startNodeOption).toBeInTheDocument();
        expect(endNodeOption).toBeInTheDocument();
        expect(addToHighValue).toBeInTheDocument();
        expect(addToOwned).toBeInTheDocument();
    });

    it('renders no asset group edit options without graph write permissions', async () => {
        server.use(
            rest.get('/api/v2/self', (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: createAuthStateWithPermissions([]).user,
                    })
                );
            })
        );

        render(<ContextMenu contextMenu={{ mouseX: 0, mouseY: 0 }} onClose={vi.fn()} />, {
            route: '/test?selectedItem=abc',
        });

        const startNodeOption = await waitFor(() => screen.findByRole('menuitem', { name: /set as starting node/i }));
        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        const addToHighValue = screen.queryByText(/add to high value/i);
        const addToOwned = screen.queryByText(/add to owned/i);

        expect(startNodeOption).toBeInTheDocument();
        expect(endNodeOption).toBeInTheDocument();
        expect(addToHighValue).toBeNull();
        expect(addToOwned).toBeNull();
    });

    it('sets a primarySearch=id and searchType=node when secondarySearch is falsey', async () => {
        render(<ContextMenu contextMenu={{ mouseX: 0, mouseY: 0 }} onClose={vi.fn()} />, {
            route: '/test?selectedItem=abc',
        });

        const startNodeOption = await waitFor(() => screen.findByRole('menuitem', { name: /set as starting node/i }));

        const user = userEvent.setup();
        await user.click(startNodeOption);

        expect(window.location.search).toContain('primarySearch=abc');
        expect(window.location.search).toContain('searchType=node');
        expect(window.location.search).toContain('exploreSearchTab=pathfinding');
    });

    it('sets a primarySearch=id and searchType=pathfinding when secondarySearch is truethy', async () => {
        render(<ContextMenu contextMenu={{ mouseX: 0, mouseY: 0 }} onClose={vi.fn()} />, {
            route: '/test?selectedItem=abc&secondarySearch=def',
        });

        const startNodeOption = await waitFor(() => screen.findByRole('menuitem', { name: /set as starting node/i }));
        const user = userEvent.setup();
        await user.click(startNodeOption);

        expect(window.location.search).toContain('primarySearch=abc');
        expect(window.location.search).toContain('searchType=pathfinding');
        expect(window.location.search).toContain('exploreSearchTab=pathfinding');
    });

    it('sets secondarySearch=id and searchType=node when primarySearch is falsey', async () => {
        render(<ContextMenu contextMenu={{ mouseX: 0, mouseY: 0 }} onClose={vi.fn()} />, {
            route: '/test?selectedItem=abc',
        });

        const endNodeOption = await waitFor(() => screen.findByRole('menuitem', { name: /set as ending node/i }));
        const user = userEvent.setup();
        await user.click(endNodeOption);

        expect(window.location.search).toContain('secondarySearch=abc');
        expect(window.location.search).toContain('searchType=node');
        expect(window.location.search).toContain('exploreSearchTab=pathfinding');
    });

    it('sets a secondary=id and searchType=pathfinding when primary is truethy', async () => {
        render(<ContextMenu contextMenu={{ mouseX: 0, mouseY: 0 }} onClose={vi.fn()} />, {
            route: '/test?selectedItem=abc&primarySearch=def',
        });

        const endNodeOption = await waitFor(() => screen.findByRole('menuitem', { name: /set as ending node/i }));
        const user = userEvent.setup();
        await user.click(endNodeOption);

        expect(window.location.search).toContain('secondarySearch=abc');
        expect(window.location.search).toContain('searchType=pathfinding');
        expect(window.location.search).toContain('exploreSearchTab=pathfinding');
    });

    it('opens a submenu when user hovers over `Copy`', async () => {
        render(<ContextMenu contextMenu={{ mouseX: 0, mouseY: 0 }} onClose={vi.fn()} />, {
            route: '/test?selectedItem=abc',
        });

        const user = userEvent.setup();

        const copyOption = await waitFor(() => screen.findByRole('menuitem', { name: /copy/i }));
        await user.hover(copyOption);

        const tip = await screen.findByRole('tooltip');
        expect(tip).toBeInTheDocument();

        const displayNameOption = screen.getByLabelText(/display name/i);
        expect(displayNameOption).toBeInTheDocument();

        const objectIdOption = screen.getByLabelText(/object id/i);
        expect(objectIdOption).toBeInTheDocument();

        const cypherOption = screen.getByLabelText(/cypher/i);
        expect(cypherOption).toBeInTheDocument();

        // hover off the `Copy` option in order to close the tooltip
        await userEvent.unhover(copyOption);

        await waitFor(() => {
            expect(screen.queryByText(/display name/i)).not.toBeInTheDocument();
            expect(screen.queryByText(/object id/i)).not.toBeInTheDocument();
            expect(screen.queryByText(/cypher/i)).not.toBeInTheDocument();
        });
    });
});
