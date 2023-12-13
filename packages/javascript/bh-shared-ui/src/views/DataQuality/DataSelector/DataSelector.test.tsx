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

import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, within } from '../../../test-utils';
import DataSelector from './';

const server = setupServer(
    rest.get(`/api/v2/available-domains`, (req, res, ctx) => {
        return res(ctx.json({ data: testDomains }));
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

const testDomains = [
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
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 73,
        name: 'antonina.info',
        id: '6b55e74d-f24e-418a-bfd1-4769e93517c7',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 53,
        name: 'orie.org',
        id: '89c928e2-79a8-4a5f-9060-d8d757cffd95',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 30,
        name: 'charley.net',
        id: '11cca0a1-db2e-4c47-b572-a8cacf8b6063',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 0,
        name: 'judson.info',
        id: '46ab17b6-6ec2-4de0-ae97-8dffb5c9a8b5',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 4,
        name: 'stanley.info',
        id: '04c01736-1575-4294-85e7-f682ec4f43ed',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 77,
        name: 'lukas.info',
        id: '0627e362-8f6a-4fce-b598-bc0454e3b094',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 53,
        name: 'kaleb.net',
        id: '9a58c28c-495a-4f07-ade2-e2e960aeeada',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 48,
        name: 'kathlyn.com',
        id: 'bb344069-17df-4626-9ed1-f2ff26f13a3f',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 88,
        name: 'sheridan.name',
        id: '82624c96-0f35-47ce-9f91-6dc58dcca87c',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 45,
        name: 'reece.net',
        id: '588e31e0-6ea3-494f-971b-744e66912754',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 81,
        name: 'zion.info',
        id: '467fcb67-f467-40fa-9a97-f98ee2d53e9f',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 52,
        name: 'fiona.org',
        id: '4a0fc37d-de3c-415c-9cc8-4dc46850f3f6',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 69,
        name: 'celine.biz',
        id: 'e2114c74-6531-497b-923d-d3321efb4dfe',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 20,
        name: 'francisca.org',
        id: 'c8367eb1-5ed7-4b20-8721-69431b99b317',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 89,
        name: 'dee.com',
        id: 'd1388a3e-59cc-4ee5-a4a1-b0e31c29fca6',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 38,
        name: 'christophe.org',
        id: '1ff0c1df-abad-4d98-8025-f6168d2a2f29',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 98,
        name: 'addison.org',
        id: '37458866-25f9-48ba-8737-2e1600959b03',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 76,
        name: 'floyd.biz',
        id: '2d3b8635-a5c3-410f-a364-891ea8878f25',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 63,
        name: 'scot.info',
        id: '66bb2d96-88bb-4209-94f3-8c64b1f84303',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 58,
        name: 'nathanial.info',
        id: '02927420-313c-444e-9031-8213f72a9807',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 82,
        name: 'idell.info',
        id: '9876f0e0-511c-4812-b749-e34f3e442c28',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 95,
        name: 'molly.org',
        id: '560c977a-a96c-4732-9aca-f98ae8f8f127',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 14,
        name: 'thora.net',
        id: '949c87fb-094a-4570-a916-e908dfae316f',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 11,
        name: 'alison.info',
        id: 'c3c7c803-ee6a-42c9-8668-dc247956087f',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 80,
        name: 'madisen.info',
        id: 'a21cb4b8-7b20-4339-99f7-f3cb11737e73',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 30,
        name: 'edmund.name',
        id: 'dd241b2d-f5aa-465f-92d6-c618489d019a',
        collected: true,
    },
    {
        type: 'azure',
        impactValue: 30,
        name: 'teagan.biz',
        id: '7b97a213-2f21-49d6-be54-92968e4701df',
        collected: true,
    },
    {
        type: 'active-directory',
        impactValue: 74,
        name: 'bartholome.biz',
        id: '4a23e881-5aee-4ca5-bea5-64bce30eaf78',
        collected: true,
    },
];

const errorMessage = <>Domains unavailable</>;

describe('Context Selector', () => {
    it('should render with a full list of multiple tenants and domains', async () => {
        const user = userEvent.setup();
        const testOnChange = vi.fn();
        const testValue = { type: 'active-directory', id: '6b55e74d-f24e-418a-bfd1-4769e93517c7' };
        render(<DataSelector value={testValue} onChange={testOnChange} errorMessage={errorMessage} />);

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
        //Both AD and Azure are available so 30 menu items are available plus both show All items for a count of 32
        expect(within(container).getAllByRole('menuitem')).toHaveLength(32);
    });

    it('should initiate data loading when an item is selected', async () => {
        const user = userEvent.setup();
        const testOnChange = vi.fn();
        const testValue = { type: 'active-directory', id: '6b55e74d-f24e-418a-bfd1-4769e93517c7' };
        render(<DataSelector value={testValue} onChange={testOnChange} errorMessage={errorMessage} />);

        const contextSelector = await screen.findByTestId('data-quality_context-selector');
        await user.click(contextSelector);

        expect(await screen.findByText(testDomains[5].name)).toBeInTheDocument();
        await user.click(screen.getByText(testDomains[5].name));

        expect(testOnChange).toHaveBeenLastCalledWith({ type: testDomains[5].type, id: testDomains[5].id });

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
        const testValue = { type: 'azure', id: 'd1993a1b-55c1-4668-9393-ddfffb6ab639' };
        render(<DataSelector value={testValue} onChange={testOnChange} errorMessage={errorMessage} />);

        const contextSelector = await screen.findByTestId('data-quality_context-selector');

        await user.click(contextSelector);

        const container = await screen.findByTestId('data-quality_context-selector-popover');
        //Only one of the served domains is collected so only one list item plus the two options for all domains and all tenants are displayed.
        expect(within(container).getAllByRole('menuitem')).toHaveLength(3);
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
        const testValue = { type: 'active-directory', id: '6b55e74d-f24e-418a-bfd1-4769e93517c7' };
        render(<DataSelector value={testValue} onChange={testOnChange} errorMessage={<>{testErrorMessage}</>} />);

        expect(await screen.findByText(testErrorMessage)).toBeInTheDocument();
    });
});
