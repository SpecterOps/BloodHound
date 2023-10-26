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

import { render, screen } from 'src/test-utils';
import CommonSearches from './CommonSearches';
import userEvent from '@testing-library/user-event';
import { setupServer } from 'msw/node';
import { rest } from 'msw';
import * as actions from 'src/ducks/explore/actions';
import { CommonSearches as prebuiltSearchList, apiClient } from 'bh-shared-ui';

const server = setupServer(
    rest.get('/api/v2/saved-queries', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        user_id: 'abcdefgh',
                        query: 'match (n) return n limit 5',
                        name: 'me save a query 1',
                        id: 1,
                    },
                    {
                        user_id: 'abcdefgh',
                        query: 'match (n) return n limit 5',
                        name: 'me save a query 2',
                        id: 2,
                    },
                ],
            })
        );
    }),
    rest.delete('/api/v2/saved-queries/:id', (req, res, ctx) => {
        return res(ctx.status(201));
    })
);

beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
    jest.restoreAllMocks();
});
afterAll(() => server.close());

describe('CommonSearches', () => {
    beforeEach(() => {
        render(<CommonSearches />);
    });

    it('renders headers', () => {
        const header = screen.getByText(/pre-built searches/i);
        const adTab = screen.getByRole('tab', { name: /active directory/i });
        const azTab = screen.getByRole('tab', { name: /azure/i });
        const userTab = screen.getByRole('tab', { name: /custom searches/i });

        expect(header).toBeInTheDocument();
        expect(adTab).toBeInTheDocument();
        expect(azTab).toBeInTheDocument();
        expect(userTab).toBeInTheDocument();

        expect(screen.getByRole('tab', { selected: true })).toHaveTextContent('Active Directory');
    });

    it('renders search list for the currently active tab', () => {
        const adSearches = prebuiltSearchList.filter(({ category }) => category === 'Active Directory');
        const subheadersForAD = adSearches.map((element) => element.subheader);

        subheadersForAD.forEach((subheader) => {
            expect(screen.getByText(subheader)).toBeInTheDocument();
        });
    });

    it('renders a different list of queries when user switches tab', async () => {
        const user = userEvent.setup();

        // switch tabs to AZ
        const azureTab = screen.getByRole('tab', { name: /azure/i });
        await user.click(azureTab);

        const azSearches = prebuiltSearchList.filter(({ category }) => category === 'Azure');
        const subheadersForAZ = azSearches.map((element) => element.subheader);

        subheadersForAZ.forEach((subheader) => {
            expect(screen.getByText(subheader)).toBeInTheDocument();
        });
    });

    it(`fetches a user's saved queries when the 'custom searches' tab is clicked`, async () => {
        const user = userEvent.setup();

        // switch tabs to user searches
        const userTab = screen.getByRole('tab', { name: /custom searches/i });
        await user.click(userTab);

        const queries = screen.getAllByRole('button', { name: /me save a query/i });
        expect(queries).toHaveLength(2);
    });

    it('handles a click on each list item', async () => {
        const spy = jest.spyOn(actions, 'startCypherQuery');
        const user = userEvent.setup();

        const adSearches = prebuiltSearchList.filter(({ category }) => category === 'Active Directory');

        const { cypher, description } = adSearches[0].queries[0];

        const listItem = screen.getByRole('button', { name: description });
        expect(listItem).toBeInTheDocument();

        await user.click(listItem);

        expect(spy).toHaveBeenCalledTimes(1);
        expect(spy).toHaveBeenCalledWith(cypher);
    });

    it('deletes a query that a user has saved', async () => {
        const spy = jest.spyOn(apiClient, 'deleteUserQuery');
        const user = userEvent.setup();

        // switch tabs to user searches
        const userTab = screen.getByRole('tab', { name: /custom searches/i });
        await user.click(userTab);

        const deleteButtons = screen.getAllByRole('button', { name: /delete query/i });
        await user.click(deleteButtons[0]);

        // verify confirmation dialog appears
        const deleteConfirmationDialog = screen.getByRole('dialog', { name: /delete query/i });
        expect(deleteConfirmationDialog).toBeInTheDocument();

        const confirmDeleteButton = screen.getByRole('button', { name: /confirm/i });
        await user.click(confirmDeleteButton);

        expect(spy).toHaveBeenCalledTimes(1);
        expect(spy).toHaveBeenCalledWith(1);
    });
});
