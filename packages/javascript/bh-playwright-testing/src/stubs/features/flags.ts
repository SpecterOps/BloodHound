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

type FeatureFlag = {
    id: number;
    key: string;
    name: string;
    description: string;
    enabled: boolean;
    user_updatable: boolean;
};

const buildEnabledFlag = (key: string): FeatureFlag => ({
    id: -1,
    key,
    name: key,
    description: `Playwright stubbed feature flag "${key}"`,
    enabled: true,
    user_updatable: false,
});

/**
 * Stubs the feature-flags endpoint (`GET /api/v2/features`) so components gated behind a flag
 * render during a test. The real response is fetched and only the flags whose `key` matches
 * `flagKeys` are forced `enabled: true`; every other flag is passed through untouched. A requested
 * key that the backend doesn't return is appended as an enabled flag so the stub is reliable
 * regardless of environment state.
 *
 * Install before navigation. Non-GET traffic falls through to any lower-priority route handlers.
 *
 * @example
 * ```
 * test.beforeEach(async ({ page, goAndWaitFor }) => {
 *     await installFeatureFlagEnabledStub(page, 'findings_table');
 *     await goAndWaitFor('/ui/graphview?view=table', page.getByText('Low').first());
 * });
 * ```
 */
export async function installFeatureFlagEnabledStub(page: Page, flagKeys: string | string[]): Promise<void> {
    const keys = Array.isArray(flagKeys) ? flagKeys : [flagKeys];

    await page.route('**/api/v2/features', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();

        const response = await route.fetch();
        const body = await response.json();

        const flags: FeatureFlag[] = Array.isArray(body?.data) ? body.data : [];
        const seen = new Set<string>();
        const patched = flags.map((flag) => {
            if (!keys.includes(flag.key)) return flag;
            seen.add(flag.key);
            return { ...flag, enabled: true };
        });

        for (const key of keys) {
            if (!seen.has(key)) patched.push(buildEnabledFlag(key));
        }

        return route.fulfill({ json: { ...body, data: patched } });
    });
}
