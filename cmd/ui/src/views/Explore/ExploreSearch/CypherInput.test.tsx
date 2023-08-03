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

import { act, fireEvent, render, screen } from 'src/test-utils';
import CypherInput from './CypherInput';
import userEvent from '@testing-library/user-event';
import * as actions from 'src/ducks/searchbar/actions';
import { CommonSearches as commonSearchesList } from 'bh-shared-ui';

describe('CypherInput', () => {
    beforeEach(async () => {
        await act(async () => {
            render(<CypherInput />);
        });
    });
    const user = userEvent.setup();

    it('should render', () => {
        expect(screen.getByPlaceholderText(/cypher search/i)).toBeInTheDocument();

        expect(screen.getByRole('button', { name: /question/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /search/i })).toBeInTheDocument();
    });

    it('a cypher search is executed when the user clicks the search button', async () => {
        vi.spyOn(actions, 'startCypherSearch');

        const cypherInput = screen.getByPlaceholderText(/cypher search/i);
        const userSuppliedCypherQuery = 'match (u:User) return u;';
        await user.type(cypherInput, userSuppliedCypherQuery);

        expect(cypherInput).toHaveTextContent(userSuppliedCypherQuery);

        const searchButton = screen.getByRole('button', { name: /search/i });
        await user.click(searchButton);

        expect(actions.startCypherSearch).toHaveBeenCalledTimes(1);
        expect(actions.startCypherSearch).toHaveBeenCalledWith(userSuppliedCypherQuery);
    });

    it('should show common cypher searches when user clicks on folder button', async () => {
        const prebuiltSearches = screen.getByText(/pre-built searches/i);
        expect(prebuiltSearches).not.toBeVisible();

        const menu = screen.getByRole('button', { name: /folder-open/i });

        await user.click(menu);
        expect(prebuiltSearches).toBeVisible();
    });

    it('when a user selects a common search, the cypher text area gets populated with the selected query', async () => {
        const commonQueryFindAllDomainAdmins = commonSearchesList[0].queries[0];
        const { description, cypher } = commonQueryFindAllDomainAdmins;

        const menu = screen.getByRole('button', { name: /folder-open/i });
        await user.click(menu);

        const firstQueryInList = screen.getByText(description);
        await user.click(firstQueryInList);

        const cypherInput = screen.getByPlaceholderText(/cypher search/i);
        expect(cypherInput).toBeInTheDocument();
        expect(cypherInput).toHaveValue(cypher);
    });

    it('a cypher search is executed when the user presses shift+enter in the text area', async () => {
        vi.spyOn(actions, 'startCypherSearch');

        const cypherInput = screen.getByPlaceholderText(/cypher search/i);
        const userSuppliedCypherQuery = 'match (u:User) return u;';
        await user.type(cypherInput, userSuppliedCypherQuery);

        expect(cypherInput).toHaveTextContent(userSuppliedCypherQuery);

        await fireEvent.keyDown(cypherInput, { key: 'Enter', shiftKey: true });

        expect(actions.startCypherSearch).toHaveBeenCalledTimes(1);
        expect(actions.startCypherSearch).toHaveBeenCalledWith(userSuppliedCypherQuery);
    });
});
