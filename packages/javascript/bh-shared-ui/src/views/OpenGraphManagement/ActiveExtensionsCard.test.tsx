// Copyright 2026 Specter Ops, Inc.
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
import { Extension } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, waitFor } from '../../test-utils';
import {
    ActiveExtensionsCard,
    ERROR_MESSAGE,
    NO_DATA_MESSAGE,
    NO_SEARCH_RESULTS_MESSAGE,
} from './ActiveExtensionsCard';

const mockExtensions: Extension[] = [
    { id: '1', name: 'Active Directory', version: 'v0.0.1' },
    { id: '2', name: 'Azure', version: 'v1.0.0' },
    { id: '3', name: 'Custom Extension', version: '0.5.0' },
];

const server = setupServer(
    rest.get(`/api/v2/extensions`, (_req, res, ctx) =>
        res(
            ctx.json({
                data: { extensions: mockExtensions },
            })
        )
    )
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('ActiveExtensionsCard', () => {
    it('renders the card with title and search input', () => {
        render(<ActiveExtensionsCard />);

        expect(screen.getByText('Active Extensions')).toBeInTheDocument();
        expect(screen.getByLabelText('Search')).toBeInTheDocument();
    });

    it('displays a loading message while fetching extensions', () => {
        render(<ActiveExtensionsCard />);

        expect(screen.getByText('Loading extensions...')).toBeInTheDocument();
    });

    it('displays an error message while fetching fails', async () => {
        server.use(
            rest.get(`/api/v2/extensions`, (req, res, ctx) => {
                return res(ctx.status(500));
            })
        );

        render(<ActiveExtensionsCard />);

        expect(await screen.findByText(ERROR_MESSAGE)).toBeInTheDocument();
    });

    it('displays no data message when there are no extensions', async () => {
        server.use(
            rest.get(`/api/v2/extensions`, (_req, res, ctx) => res.once(ctx.json({ data: { extensions: [] } })))
        );

        render(<ActiveExtensionsCard />);

        expect(await screen.findByText(NO_DATA_MESSAGE)).toBeInTheDocument();
    });

    it('displays extensions in a table when data is available', async () => {
        render(<ActiveExtensionsCard />);

        expect(await screen.findByText('Active Directory')).toBeInTheDocument();
        expect(screen.getByText('v0.0.1')).toBeInTheDocument();
        expect(screen.getByText('Azure')).toBeInTheDocument();
        expect(screen.getByText('v1.0.0')).toBeInTheDocument();
        expect(screen.getByText('Custom Extension')).toBeInTheDocument();
        expect(screen.getByText('0.5.0')).toBeInTheDocument();
    });

    it('renders delete buttons for each extension', async () => {
        render(<ActiveExtensionsCard />);

        expect(await screen.findByLabelText('Delete Active Directory')).toBeInTheDocument();
        expect(screen.getByLabelText('Delete Azure')).toBeInTheDocument();
        expect(screen.getByLabelText('Delete Custom Extension')).toBeInTheDocument();
    });

    it('filters extensions based on search input', async () => {
        const user = userEvent.setup();

        render(<ActiveExtensionsCard />);

        const searchInput = screen.getByLabelText('Search');
        await user.type(searchInput, 'Azure');

        await waitFor(() => {
            expect(screen.getByText('Azure')).toBeInTheDocument();
            expect(screen.queryByText('Active Directory')).not.toBeInTheDocument();
            expect(screen.queryByText('Custom Extension')).not.toBeInTheDocument();
        });
    });

    it('performs case-insensitive search', async () => {
        const user = userEvent.setup();

        render(<ActiveExtensionsCard />);

        const searchInput = screen.getByLabelText('Search');
        await user.type(searchInput, 'azure');

        await waitFor(() => {
            expect(screen.getByText('Azure')).toBeInTheDocument();
            expect(screen.queryByText('Active Directory')).not.toBeInTheDocument();
        });
    });

    it('displays no search results message when search yields no matches', async () => {
        const user = userEvent.setup();

        render(<ActiveExtensionsCard />);

        const searchInput = screen.getByLabelText('Search');
        await user.type(searchInput, 'NonExistent');

        await waitFor(() => {
            expect(screen.getByText(NO_SEARCH_RESULTS_MESSAGE)).toBeInTheDocument();
        });
    });

    it('shows all extensions when search is cleared', async () => {
        const user = userEvent.setup();

        render(<ActiveExtensionsCard />);

        const searchInput = screen.getByLabelText('Search');
        await user.type(searchInput, 'Azure');

        await waitFor(() => {
            expect(screen.getByText('Azure')).toBeInTheDocument();
            expect(screen.queryByText('Custom Extension')).not.toBeInTheDocument();
        });

        await user.clear(searchInput);

        await waitFor(() => {
            expect(screen.getByText('Azure')).toBeInTheDocument();
            expect(screen.getByText('Active Directory')).toBeInTheDocument();
            expect(screen.getByText('Custom Extension')).toBeInTheDocument();
        });
    });

    it('applies correct dynamic height based on filtered data', async () => {
        const { container } = render(<ActiveExtensionsCard />);

        await screen.findByText('Active Directory');

        const tableContainer = container.querySelector('div[style*="min-height"]');
        expect(tableContainer).toBeInTheDocument();
        // With 3 extensions: TABLE_HEADER_HEIGHT (52) + TABLE_CELL_HEIGHT (57) * 3 = 223px
        expect(tableContainer).toHaveStyle({ minHeight: '223px' });
    });

    it('applies empty state height when no results', async () => {
        server.use(
            rest.get(`/api/v2/extensions`, (_req, res, ctx) => res.once(ctx.json({ data: { extensions: [] } })))
        );

        const { container } = render(<ActiveExtensionsCard />);

        await screen.findByText(NO_DATA_MESSAGE);

        const tableContainer = container.querySelector('div[style*="min-height"]');
        expect(tableContainer).toBeInTheDocument();
    });
});
