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

export const GRAPH_HAS_DATA_QUERY = 'MATCH (A) WHERE NOT A:MigrationData RETURN A LIMIT 1';

const GRAPH_HAS_DATA_RESPONSE = {
    data: {
        nodes: {
            seed: {
                isOwnedObject: false,
                isTierZero: false,
                kind: 'Group',
                label: 'PLAYWRIGHT_SEED',
                objectId: 'playwright-seed',
            },
        },
        edges: {},
    },
};

const GRAPH_HAS_NO_DATA_RESPONSE = {
    data: {
        nodes: {},
        edges: {},
    },
};

// Stub every POST to the cypher endpoint with a populated payload so the `useGraphHasData`
// probe resolves to "true" and the "No Data Available" upload dialog never settles open.
// Non-POST traffic falls through.
//
// Install before navigation. Tests that need a different cypher response can register a
// higher-priority `page.route` for the same URL — Playwright runs handlers in LIFO order,
// so a test-local handler wins for the cases it cares about.
export async function installGraphHasDataStub(page: Page): Promise<void> {
    await page.route('**/api/v2/graphs/cypher', async (route) => {
        if (route.request().method() !== 'POST') {
            return route.fallback();
        }

        return route.fulfill({
            json: GRAPH_HAS_DATA_RESPONSE,
        });
    });
}

// Overrides only the `useGraphHasData` probe with an empty payload so the "No Data Available"
// upload dialog renders. Other cypher requests fall through to any lower-priority route handlers.
export async function installGraphHasNoDataStub(page: Page): Promise<void> {
    await page.route('**/api/v2/graphs/cypher', async (route) => {
        const body = route.request().postDataJSON();
        if (body?.query === GRAPH_HAS_DATA_QUERY) {
            return route.fulfill({
                json: GRAPH_HAS_NO_DATA_RESPONSE,
            });
        }

        return route.fallback();
    });
}
