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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { SimpleEnvironmentSelector } from '.';
import { testEnvironments } from '../../mocks/handlers/environments';
import { render, screen, within } from '../../test-utils';

const server = setupServer(
    rest.get(`/api/v2/available-domains`, (req, res, ctx) => {
        return res(ctx.json({ data: testEnvironments }));
    }),
    rest.get(`/api/v2/features`, (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        id: 1,
                        key: 'azure_support',
                        name: 'azure_support',
                        description: 'azure_support',
                        enabled: true,
                        user_updatable: true,
                    },
                ],
            })
        );
    })
);
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const errorMessage = <>Domains unavailable</>;

describe('Simple Environment Selector', () => {
    it('should render with a full list of multiple tenants and domains', async () => {
        const user = userEvent.setup();
        const testOnChange = vi.fn();
        const testValue = { type: 'active-directory', id: '3a6f8001-11f4-43bb-9de6-25c0d931f244' } as const;

        render(<SimpleEnvironmentSelector selected={testValue} onSelect={testOnChange} errorMessage={errorMessage} />);

        const contextSelector = await screen.findByTestId('data-quality_context-selector');
        expect(contextSelector).toBeInTheDocument();

        expect(screen.queryByLabelText('Search')).toBeNull();

        await user.click(contextSelector);

        const searchInput = await screen.findByTestId('data-quality_context-selector-search');
        expect(searchInput).toBeInTheDocument();
        expect(searchInput).toBeVisible();

        expect(await screen.findByText('All Active Directory Domains')).toBeInTheDocument();
        expect(await screen.findByText('All Azure Tenants')).toBeInTheDocument();

        const container = await screen.findByTestId('data-quality_context-selector-popover');
        // With default filter:
        // Azure - 7
        // Active Directory - 1
        // Plus 1 for each aggregate ('All Azure', 'All Active Directory')
        expect(within(container).getAllByRole('button')).toHaveLength(10);
    });

    it('should initiate data loading when an item is selected', async () => {
        const user = userEvent.setup();
        const testOnChange = vi.fn();
        const testValue = { type: 'active-directory', id: '6b55e74d-f24e-418a-bfd1-4769e93517c7' } as const;
        render(<SimpleEnvironmentSelector selected={testValue} onSelect={testOnChange} errorMessage={errorMessage} />);

        const contextSelector = await screen.findByTestId('data-quality_context-selector');
        await user.click(contextSelector);

        expect(await screen.findByText(testEnvironments[5].name)).toBeInTheDocument();
        await user.click(screen.getByText(testEnvironments[5].name));

        expect(testOnChange).toHaveBeenLastCalledWith({ type: testEnvironments[5].type, id: testEnvironments[5].id });

        await user.click(contextSelector);
        await user.click(screen.getByText('All Active Directory Domains'));

        expect(testOnChange).toHaveBeenLastCalledWith({ type: 'active-directory-platform', id: null });
    });
});

describe('Context Selector', () => {
    beforeEach(() => {
        server.use(
            rest.get(`/api/v2/available-domains`, (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: [
                            {
                                type: 'azure',
                                impactValue: 95,
                                name: 'talia.info',
                                id: 'd1993a1b-55c1-4668-9393-ddfffb6ab639',
                                collected: true,
                            },
                            {
                                type: 'azure',
                                impactValue: 44,
                                name: 'ryleigh.com',
                                id: '844b8b80-db82-473d-8f0e-4bab2034c2da',
                                collected: false,
                            },
                        ],
                    })
                );
            })
        );
    });

    it('should not render list items for domains that are not collected', async () => {
        const user = userEvent.setup();
        const testOnChange = vi.fn();
        const testValue = { type: 'azure', id: 'd1993a1b-55c1-4668-9393-ddfffb6ab639' } as const;

        render(<SimpleEnvironmentSelector selected={testValue} onSelect={testOnChange} errorMessage={errorMessage} />);

        const contextSelector = await screen.findByTestId('data-quality_context-selector');

        await user.click(contextSelector);

        const container = await screen.findByTestId('data-quality_context-selector-popover');

        //Only one of the served domains is collected so only one list item plus the one option for all tenants are displayed.
        expect(within(container).getAllByRole('button')).toHaveLength(2);
    });
});

describe('Context Selector Error', () => {
    beforeEach(() => {
        console.error = vi.fn();
        server.use(
            rest.get(`/api/v2/available-domains`, (req, res, ctx) => {
                return res(
                    ctx.status(500),
                    ctx.json({
                        errorMessage: `Internal Server Error`,
                    })
                );
            })
        );
    });

    it('should display an error message if data does not return from the API', async () => {
        const testOnChange = vi.fn();
        const testErrorMessage = 'test error message';
        const testValue = { type: 'active-directory', id: '6b55e74d-f24e-418a-bfd1-4769e93517c7' } as const;
        render(
            <SimpleEnvironmentSelector
                selected={testValue}
                onSelect={testOnChange}
                errorMessage={<>{testErrorMessage}</>}
            />
        );

        expect(await screen.findByText(testErrorMessage)).toBeInTheDocument();
    });
});
