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

import type { Page } from '@playwright/test';
import { expectNoAccessibilityViolations, test } from '../../fixtures';

const SEARCH_TERM = 'test';
const OBJECT_ID = 'playwright-gpo-1';
const GRAPH_ID = '42';
const OU_GRAPH_ID = '43';
const GPO_NAME = 'TEST_GPO@PLAYWRIGHT.LOCAL';
const RELATIONSHIP_COUNT = 1000;
const SELECTABLE_OU_NAME = 'DOMAIN CONTROLLERS@PLAYWRIGHT.LOCAL';

const selectedNode = {
    label: GPO_NAME,
    kind: 'GPO',
    kinds: ['GPO'],
    objectId: OBJECT_ID,
    isTierZero: false,
    isOwnedObject: false,
    lastSeen: '',
};

const createRelationshipResponse = (count: number, data: Array<Record<string, unknown>>) => ({
    count,
    skip: 0,
    limit: 128,
    data,
});

const mockCommonRoutes = async (page: Page) => {
    await page.route('**/api/v2/search*', async (route) => {
        if (route.request().method() !== 'GET') {
            return route.fallback();
        }

        return route.fulfill({
            json: {
                data: [{ name: GPO_NAME, objectid: OBJECT_ID, type: 'GPO' }],
            },
        });
    });

    await page.route('**/api/v2/graph-search*', async (route) => {
        return route.fulfill({
            json: {
                data: {
                    [GRAPH_ID]: {
                        data: {
                            objectid: OBJECT_ID,
                            name: GPO_NAME,
                            nodetype: 'GPO',
                            kinds: ['GPO'],
                        },
                        label: {
                            text: GPO_NAME,
                        },
                    },
                },
            },
        });
    });

    await page.route(`**/api/v2/nodes/${GRAPH_ID}*`, async (route) => {
        const request = route.request();
        const url = new URL(request.url());

        if (request.method() !== 'GET' || url.searchParams.get('include-info') !== 'true') {
            return route.fallback();
        }

        return route.fulfill({
            json: {
                data: {
                    node_id: Number(GRAPH_ID),
                    kinds: [{ name: 'GPO', node_kind_id: 1 }],
                    properties: {
                        objectid: OBJECT_ID,
                        name: GPO_NAME,
                    },
                },
            },
        });
    });

    await page.route('**/api/v2/graphs/source-kinds', async (route) => {
        return route.fulfill({
            json: {
                data: {
                    kinds: [],
                },
            },
        });
    });

    await page.route('**/api/v2/features', async (route) => {
        return route.fulfill({
            json: {
                data: [
                    {
                        id: 1,
                        key: 'tier_management_engine',
                        name: 'Tier Management Engine',
                        description: '',
                        enabled: false,
                        user_updatable: false,
                    },
                ],
            },
        });
    });
};

const selectMockedGPO = async (page: Page) => {
    await page.goto('/ui/explore');

    const searchInput = page.getByLabel('Search Nodes');
    await searchInput.waitFor({ state: 'visible' });
    await searchInput.click();
    await searchInput.pressSequentially(SEARCH_TERM);

    const searchResult = page.getByTestId('explore_search_result-list-item').first();
    await searchResult.waitFor({ state: 'visible' });
    await searchResult.click();
};

test.describe('WCAG A/AA Violations - Explore - Entity Information Panel', () => {
    test('With a relationship section expanded', async ({ page, makeAxeBuilder }, testInfo) => {
        await mockCommonRoutes(page);

        await page.route(`**/api/v2/gpos/${OBJECT_ID}/*`, async (route) => {
            return route.fulfill({
                json: createRelationshipResponse(
                    RELATIONSHIP_COUNT,
                    Array.from({ length: 128 }, (_, index) => ({
                        kind: 'User',
                        props: {
                            name: `RELATED_USER_${index}@PLAYWRIGHT.LOCAL`,
                            objectid: `playwright-related-user-${index}`,
                        },
                    }))
                ),
            });
        });

        await selectMockedGPO(page);

        const infoPanel = page.getByTestId('explore_entity-information-panel');
        await infoPanel.waitFor({ state: 'visible' });

        await infoPanel.getByTestId('entity-object-information-skeleton').waitFor({ state: 'detached' });

        const inboundObjectControlSection = infoPanel.getByRole('button', {
            name: /Inbound Object Control/,
        });
        await inboundObjectControlSection.waitFor({ state: 'visible' });
        await inboundObjectControlSection.click();

        await infoPanel.getByText('RELATED_USER_0@PLAYWRIGHT.LOCAL').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With disabled sections', async ({ page, makeAxeBuilder }, testInfo) => {
        await mockCommonRoutes(page);

        await page.route(`**/api/v2/gpos/${OBJECT_ID}/*`, async (route) => {
            const url = new URL(route.request().url());

            if (url.pathname.endsWith('/controllers')) {
                return route.fulfill({
                    json: createRelationshipResponse(0, []),
                });
            }

            return route.fulfill({
                json: createRelationshipResponse(
                    RELATIONSHIP_COUNT,
                    Array.from({ length: 128 }, (_, index) => ({
                        kind: 'User',
                        props: {
                            name: `RELATED_USER_${index}@PLAYWRIGHT.LOCAL`,
                            objectid: `playwright-related-user-${index}`,
                        },
                    }))
                ),
            });
        });

        await selectMockedGPO(page);

        const infoPanel = page.getByTestId('explore_entity-information-panel');
        await infoPanel.waitFor({ state: 'visible' });
        await infoPanel.getByTestId('entity-object-information-skeleton').waitFor({ state: 'detached' });

        await infoPanel.getByRole('button', { name: /Inbound Object Control/ }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With selectable items section', async ({ page, makeAxeBuilder }, testInfo) => {
        await mockCommonRoutes(page);

        await page.route(`**/api/v2/gpos/${OBJECT_ID}/*`, async (route) => {
            const url = new URL(route.request().url());

            if (url.pathname.endsWith('/ous')) {
                if (url.searchParams.get('type') === 'graph') {
                    return route.fulfill({
                        json: {
                            [OU_GRAPH_ID]: {
                                color: '#000000',

                                data: {
                                    nodetype: 'OU',
                                    kinds: ['Base', 'OU', 'Tag_Tier_Zero'],
                                    objectid: 'playwright-ou-1',
                                    name: SELECTABLE_OU_NAME,
                                    system_tags: [],
                                },
                                border: {
                                    color: '#000000',
                                },
                                fontIcon: {
                                    text: '',
                                },
                                label: {
                                    backgroundColor: '#ffffff',
                                    center: true,
                                    fontSize: 12,
                                    text: SELECTABLE_OU_NAME,
                                },
                                size: 1,
                            },
                        },
                    });
                }

                return route.fulfill({
                    json: createRelationshipResponse(1, [
                        {
                            kind: 'OU',
                            props: {
                                name: SELECTABLE_OU_NAME,
                                objectid: 'playwright-ou-1',
                            },
                        },
                    ]),
                });
            }

            if (
                url.pathname.endsWith('/computers') ||
                url.pathname.endsWith('/users') ||
                url.pathname.endsWith('/tier-zero')
            ) {
                return route.fulfill({
                    json: createRelationshipResponse(0, []),
                });
            }

            return route.fulfill({
                json: createRelationshipResponse(0, []),
            });
        });

        await selectMockedGPO(page);

        const infoPanel = page.getByTestId('explore_entity-information-panel');
        await infoPanel.waitFor({ state: 'visible' });
        await infoPanel.getByTestId('entity-object-information-skeleton').waitFor({ state: 'detached' });

        const affectedObjectsSection = infoPanel.getByRole('button', { name: /Affected Objects/ });
        await affectedObjectsSection.waitFor({ state: 'visible' });
        await affectedObjectsSection.click();

        const ousSection = infoPanel.getByRole('button', { name: /OUs/ });
        await ousSection.waitFor({ state: 'visible' });
        await ousSection.click();

        await infoPanel.getByText(SELECTABLE_OU_NAME).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
