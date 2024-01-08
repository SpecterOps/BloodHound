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

import { act } from 'react-dom/test-utils';
import { render, screen, waitFor } from 'src/test-utils';
import userEvent from '@testing-library/user-event';
import ContextMenu from './ContextMenu';
import * as actions from 'src/ducks/searchbar/actions';
import { setupServer } from 'msw/node';
import { rest } from 'msw';

describe('ContextMenu', async () => {
    const server = setupServer(
        rest.get('/api/v2/asset-groups/:assetGroupId/members', (req, res, ctx) => {
            return res(
                ctx.json({
                    data: {
                        members: [],
                    },
                })
            );
        })
    );

    beforeAll(() => server.listen());
    beforeEach(async () => {
        await act(async () => {
            render(<ContextMenu contextMenu={{ mouseX: 0, mouseY: 0 }} handleClose={vi.fn()} />, {
                initialState: {
                    entityinfo: {
                        selectedNode: {
                            name: 'foo',
                            id: '1234',
                            type: 'User',
                        },
                    },
                    assetgroups: {
                        assetGroups: [
                            { tag: 'owned', id: 1 },
                            { tag: 'admin_tier_0', id: 2 },
                        ],
                    },
                },
            });
        });
    });
    afterEach(() => {
        server.resetHandlers();
    });
    afterAll(() => server.close());

    it('renders', () => {
        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        const addToHighValueOption = screen.getByRole('menuitem', { name: /add to high value/i });
        const addToOwnedOption = screen.getByRole('menuitem', { name: /add to owned/i });

        expect(startNodeOption).toBeInTheDocument();
        expect(endNodeOption).toBeInTheDocument();
        expect(addToHighValueOption).toBeInTheDocument();
        expect(addToOwnedOption).toBeInTheDocument();
    });

    it('handles setting a start node', async () => {
        const user = userEvent.setup();
        const sourceNodeSelectedSpy = vi.spyOn(actions, 'sourceNodeSelected');

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        await user.click(startNodeOption);

        expect(sourceNodeSelectedSpy).toBeCalledTimes(1);
        expect(sourceNodeSelectedSpy).toHaveBeenCalledWith(
            {
                name: 'foo',
                objectid: '1234',
                type: 'User',
            },
            true
        );
    });

    it('handles setting an end node', async () => {
        const user = userEvent.setup();
        const destinationNodeSelectedSpy = vi.spyOn(actions, 'destinationNodeSelected');

        const startNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        await user.click(startNodeOption);

        expect(destinationNodeSelectedSpy).toBeCalledTimes(1);
        expect(destinationNodeSelectedSpy).toHaveBeenCalledWith({
            name: 'foo',
            objectid: '1234',
            type: 'User',
        });
    });

    it('opens a submenu when user hovers over `Copy`', async () => {
        const user = userEvent.setup();

        const copyOption = screen.getByRole('menuitem', { name: /copy/i });
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
