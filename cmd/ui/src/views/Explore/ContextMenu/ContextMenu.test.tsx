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
import { render, screen } from 'src/test-utils';
import userEvent from '@testing-library/user-event';
import ContextMenu from './ContextMenu';
import * as actions from 'src/ducks/searchbar/actions';
import { PRIMARY_SEARCH, SEARCH_TYPE_EXACT, PATHFINDING_SEARCH } from 'src/ducks/searchbar/types';

describe('ContextMenu', async () => {
    beforeEach(async () => {
        await act(async () => {
            render(<ContextMenu anchorPosition={{ x: 0, y: 0 }} />, {
                initialState: {
                    entityinfo: {
                        selectedNode: {
                            name: 'foo',
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
        const sourceNodeSuggestedSpy = vi.spyOn(actions, 'sourceNodeSuggested');

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        await user.click(startNodeOption);

        expect(sourceNodeSuggestedSpy).toBeCalledTimes(1);
        expect(sourceNodeSuggestedSpy).toHaveBeenCalledWith('foo');
    });

    it('handles setting a end node', async () => {
        const user = userEvent.setup();
        const setActiveTabSpy = vi.spyOn(actions, 'tabChanged');
        const setSearchValueSpy = vi.spyOn(actions, 'setSearchValue');
        const startSearchActionSpy = vi.spyOn(actions, 'startSearchAction');

        const startNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        await user.click(startNodeOption);

        expect(setActiveTabSpy).toBeCalledTimes(1);
        expect(setActiveTabSpy).toHaveBeenCalledWith(PATHFINDING_SEARCH);

        expect(setSearchValueSpy).toBeCalledTimes(1);
        expect(setSearchValueSpy).toHaveBeenCalledWith(null, PATHFINDING_SEARCH, SEARCH_TYPE_EXACT);

        expect(startSearchActionSpy).toBeCalledTimes(1);
        expect(startSearchActionSpy).toHaveBeenCalledWith('foo', PATHFINDING_SEARCH);
    });
});
