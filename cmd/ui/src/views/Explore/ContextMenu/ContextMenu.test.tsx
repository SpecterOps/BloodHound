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

describe('ContextMenu', async () => {
    beforeEach(async () => {
        await act(async () => {
            render(<ContextMenu anchorPosition={{ x: 0, y: 0 }} />, {
                initialState: {
                    entityinfo: {
                        selectedNode: {
                            name: 'foo',
                            id: '1234',
                            type: 'User',
                        },
                    },
                },
            });
        });
    });

    it('renders', () => {
        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });

        expect(startNodeOption).toBeInTheDocument();
        expect(endNodeOption).toBeInTheDocument();
    });

    it('handles setting a start node', async () => {
        const user = userEvent.setup();
        const sourceNodeSelectedSpy = vi.spyOn(actions, 'sourceNodeSelected');

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        await user.click(startNodeOption);

        expect(sourceNodeSelectedSpy).toBeCalledTimes(1);
        expect(sourceNodeSelectedSpy).toHaveBeenCalledWith({
            name: 'foo',
            objectid: '1234',
            type: 'User',
        });
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

    it('opens a submenu when user clicks `Copy`', async () => {
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

    it('handles copying a display name', async () => {
        const user = userEvent.setup();

        const copyOption = screen.getByRole('menuitem', { name: /copy/i });
        await user.click(copyOption);

        // the tooltip container and the menu item for `display name` have the same accesible name, so return the second element here (which is the menu item)
        const displayNameOption = screen.getAllByRole('menuitem', { name: /display name/i })[1];
        await user.click(displayNameOption);

        const clipboardText = await navigator.clipboard.readText();
        expect(clipboardText).toBe('foo');
    });
});
