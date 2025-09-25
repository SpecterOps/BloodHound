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

import { render } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { QueryClient, QueryClientProvider } from 'react-query';
import { vi } from 'vitest';
import CommonSearches from './CommonSearches';

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
    }),
    rest.get('/api/v2/features', async (req, res, ctx) => {
        return res(
            ctx.json({
                data: [{ id: 16, key: 'tier_management_engine', enabled: true }],
            })
        );
    }),
    rest.get('/api/v2/self', async (req, res, ctx) => {
        return res(
            ctx.json({
                data: { id: '4e09c965-65bd-4f15-ae71-5075a6fed14b' },
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
    vi.restoreAllMocks();
});
afterAll(() => server.close());

const queryClient = new QueryClient();

describe('CommonSearches', () => {
    it('renders headers', async () => {
        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={true}
                />
            </QueryClientProvider>
        );

        const header = screen.getByText(/Saved Queries/i);
        expect(header).toBeInTheDocument();
    });

    it('should display filter dropwdowns', async () => {
        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={true}
                />
            </QueryClientProvider>
        );

        const platformLabel = await screen.findByLabelText(/platforms/i);
        const categoriesLabel = await screen.findByLabelText(/categories/i);
        const sourceLabel = await screen.findByLabelText(/source/i);
        expect(platformLabel).toBeInTheDocument();
        expect(categoriesLabel).toBeInTheDocument();
        expect(sourceLabel).toBeInTheDocument();
    });

    it('renders a filter search and platform dropdown menu', async () => {
        const user = userEvent.setup();

        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={false}
                />
            </QueryClientProvider>
        );

        const testSearch = await screen.findByPlaceholderText('Search');
        expect(testSearch).toBeInTheDocument();
        expect(testSearch).toHaveValue('');
        const testPlatforms = await screen.findByLabelText(/platform/i);
        expect(testPlatforms).toBeInTheDocument();
        await user.click(testPlatforms);
        const testListBox = await screen.findByRole('listbox');
        expect(testListBox).toBeInTheDocument();
        expect(testListBox).toBeVisible();

        const ulElement = testListBox;
        expect(ulElement.children).toHaveLength(4);

        await user.click(ulElement.children[0]);

        expect(screen.getByText(/all domain admins/i)).toBeInTheDocument();
    });

    it('displays correct content based on platform filter Azure', async () => {
        const user = userEvent.setup();

        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={false}
                />
            </QueryClientProvider>
        );

        const testPlatforms = await screen.findByLabelText(/platform/i);
        expect(testPlatforms).toBeInTheDocument();
        await user.click(testPlatforms);
        const testListBox = await screen.findByRole('listbox');
        expect(testListBox).toBeInTheDocument();
        expect(testListBox).toBeVisible();

        const ulElement = testListBox;
        expect(ulElement.children).toHaveLength(4);

        //select Azure
        await user.click(ulElement.children[2]);

        //Azure query present
        expect(screen.getByText(/All members of high privileged roles/i)).toBeInTheDocument();

        //AD query not present
        const adText = screen.queryByText(/all domain admins/i);
        expect(adText).toBeNull();
    });

    it('displays correct content based on platform filter AD', async () => {
        const user = userEvent.setup();

        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={false}
                />
            </QueryClientProvider>
        );

        const testPlatforms = await screen.findByLabelText(/platform/i);
        expect(testPlatforms).toBeInTheDocument();
        await user.click(testPlatforms);
        const testListBox = await screen.findByRole('listbox');
        const ulElement = testListBox;
        expect(ulElement.children).toHaveLength(4);

        //select AD
        await user.click(ulElement.children[1]);

        //AD query present
        expect(screen.getByText(/all domain admins/i)).toBeInTheDocument();

        //Axure query not present
        const adText = screen.queryByText(/All members of high privileged roles/i);
        expect(adText).toBeNull();
    });

    //Toggle switch - test visibility
    it('handles chevron click event', async () => {
        const user = userEvent.setup();
        const handleToggle = vi.fn();
        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={handleToggle}
                    showCommonQueries={true}
                />
            </QueryClientProvider>
        );
        const queriesToggle = screen.getByTestId('common-queries-toggle');
        await user.click(queriesToggle);
        expect(handleToggle).toBeCalled();
        expect(handleToggle).toBeCalledTimes(1);
        expect(screen.getByText(/chevron-down/i)).toBeInTheDocument();
    });
});
