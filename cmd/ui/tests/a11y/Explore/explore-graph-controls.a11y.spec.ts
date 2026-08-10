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
import type { StyledGraphEdge, StyledGraphNode } from 'js-client-library';
import { expectNoAccessibilityViolations, test } from '../../fixtures';

const CYPHER_QUERY = 'MATCH (n) RETURN n LIMIT 2';
const EXPLORE_URL =
    '/ui/explore?exploreSearchTab=cypher&searchType=cypher&cypherSearch=TUFUQ0ggKG4pIFJFVFVSTiBuIExJTUlUIDI%3D';
const FIRST_NODE_GRAPH_ID = 'data';
const SECOND_NODE_GRAPH_ID = 'playwright-node-2';
const EDGE_GRAPH_ID = 'playwright-edge-1';
const FIRST_NODE_OBJECT_ID = 'playwright-object-1';
const SECOND_NODE_OBJECT_ID = 'playwright-object-2';
const FIRST_NODE_LABEL = 'PLAYWRIGHT GRAPH USER';
const SECOND_NODE_LABEL = 'PLAYWRIGHT GRAPH GROUP';

const createStyledNode = (objectId: string, label: string, nodeType: string): StyledGraphNode => ({
    color: '#5c6bc0',
    data: {
        kinds: ['Base', nodeType],
        lastseen: '2026-01-01T00:00:00Z',
        name: label,
        nodetype: nodeType,
        objectid: objectId,
        system_tags: [],
    },
    border: {
        color: '#1a237e',
    },
    fontIcon: {
        text: '',
    },
    label: {
        backgroundColor: '#ffffff',
        center: true,
        fontSize: 12,
        text: label,
    },
    size: 1,
});

const firstNode = {
    ...createStyledNode(FIRST_NODE_OBJECT_ID, FIRST_NODE_LABEL, 'User'),
    // The Cypher query checks this legacy response location before the flat graph is normalized.
    nodes: { [FIRST_NODE_GRAPH_ID]: true },
};
const secondNode = createStyledNode(SECOND_NODE_OBJECT_ID, SECOND_NODE_LABEL, 'Group');
const edge: StyledGraphEdge = {
    id: 1,
    color: '#9e9e9e',
    data: {
        composite_risk_impact_percent: 25,
        lastseen: '2026-01-01T00:00:00Z',
    },
    end1: {
        arrow: false,
    },
    end2: {
        arrow: true,
    },
    id1: FIRST_NODE_GRAPH_ID,
    id2: SECOND_NODE_GRAPH_ID,
    label: {
        text: 'MemberOf',
    },
};

const installExploreGraphRoute = async (page: Page) => {
    await page.route('**/api/v2/graphs/cypher', async (route) => {
        const request = route.request();

        if (request.method() !== 'POST' || request.postDataJSON()?.query !== CYPHER_QUERY) {
            return route.fallback();
        }

        return route.fulfill({
            json: {
                [FIRST_NODE_GRAPH_ID]: firstNode,
                [SECOND_NODE_GRAPH_ID]: secondNode,
                [EDGE_GRAPH_ID]: edge,
            },
        });
    });
};

const loadExploreGraph = async (page: Page) => {
    await installExploreGraphRoute(page);
    await page.goto(EXPLORE_URL);
    await page.getByTestId('sigma-container-wrapper').waitFor({ state: 'attached' });
    await page.getByTestId('explore_graph-controls').waitFor({ state: 'visible' });
};

const expandGraphSearch = async (page: Page) => {
    const searchButton = page.getByRole('button', {
        name: 'Search node in results',
        exact: true,
    });

    await searchButton.click();

    const searchInput = page.getByPlaceholder('Search node in results');
    await searchInput.waitFor({ state: 'visible' });

    return searchInput;
};

test.describe('WCAG A/AA Violations - Explore - Graph Controls', () => {
    test.beforeEach(async ({ page }) => {
        await loadExploreGraph(page);
    });

    test('With Hide Labels expanded', async ({ page, makeAxeBuilder }, testInfo) => {
        const hideLabelsButton = page.getByRole('button', { name: 'Hide Labels', exact: true });
        await hideLabelsButton.click();

        const hideAllLabelsMenuItem = page.getByRole('menuitem', {
            name: 'Hide All Labels Toggle',
            exact: true,
        });
        await hideAllLabelsMenuItem.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="menu"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With Layout expanded', async ({ page, makeAxeBuilder }, testInfo) => {
        const layoutButton = page.getByRole('button', { name: 'Layout', exact: true });
        await layoutButton.click();

        const sequentialMenuItem = page.getByRole('menuitem', { name: 'Sequential', exact: true });
        await sequentialMenuItem.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="menu"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With Export expanded', async ({ page, makeAxeBuilder }, testInfo) => {
        const exportButton = page.getByRole('button', { name: 'Export', exact: true });
        await exportButton.click();

        const jsonMenuItem = page.getByRole('menuitem', { name: 'JSON', exact: true });
        await jsonMenuItem.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="menu"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With Search expanded', async ({ page, makeAxeBuilder }, testInfo) => {
        await expandGraphSearch(page);

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With search showing results', async ({ page, makeAxeBuilder }, testInfo) => {
        const searchInput = await expandGraphSearch(page);
        await searchInput.fill(FIRST_NODE_LABEL);

        const matchingResult = page.getByRole('option').filter({ hasText: FIRST_NODE_LABEL });
        await matchingResult.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With search showing no results', async ({ page, makeAxeBuilder }, testInfo) => {
        const searchInput = await expandGraphSearch(page);
        await searchInput.fill('no-matching-playwright-node');

        const noResultsMessage = page.getByText('No result found in current results', { exact: true });
        await noResultsMessage.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
