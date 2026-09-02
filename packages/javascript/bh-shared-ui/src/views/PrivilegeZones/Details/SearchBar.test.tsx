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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../../test-utils';
import SearchBar from './SearchBar';

const mockResults = {
    data: {
        data: {
            tags: [{ id: 1, name: 'Test Zone A' }],
            selectors: [{ id: 123, name: 'Test Rule A', asset_group_tag_id: 1 }],
            members: [{ id: 456, name: 'Test Member A', asset_group_tag_id: 1 }],
        },
    },
};

const server = setupServer(
    rest.post('/api/v2/asset-group-tags/search', async (_req, res, ctx) => {
        return res(ctx.json(mockResults));
    })
);

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('SearchBar', () => {
    it('renders an input box', () => {
        render(<SearchBar />);
        expect(screen.getByTestId('privilege-zone-detail-search-bar')).toBeInTheDocument();
    });

    it('Gets focus on keyboard shortcut', async () => {
        render(<SearchBar />);
        const user = userEvent.setup();
        expect(screen.getByTestId('privilege-zone-detail-search-bar')).not.toHaveFocus();

        await user.keyboard('{Alt>}{Shift>}[Slash]{/Shift}{/Alt}');

        expect(screen.getByTestId('privilege-zone-detail-search-bar')).toHaveFocus();
    });

    it('does not trigger search for fewer than 3 characters', async () => {
        render(<SearchBar />);

        const input = screen.getByPlaceholderText(/search/i);
        const user = userEvent.setup();
        await user.type(input, 'ab');

        // Nothing should appear
        expect(screen.queryByText('Zones')).not.toBeInTheDocument();
    });
});
