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

import { expect, expectNoAccessibilityViolations, test } from 'bh-playwright-testing';

test.describe('Login page accessibility', () => {
    // The auth setup project snapshots a logged-in storageState for all other a11y specs, but
    // the login page only renders for unauthenticated users — with auth present, it redirects
    // to ROUTE_HOME. Null out `persistedState.auth.sessionToken` (which lives in localStorage,
    // not cookies — see cmd/ui/src/store.ts) before any page script runs, so the app boots
    // into the unauthenticated state. We deliberately keep the rest of `persistedState`
    // intact so dark-mode styling still applies when this spec runs under the *-dark project.
    test.beforeEach(async ({ context }) => {
        await context.addInitScript(() => {
            try {
                const raw = window.localStorage.getItem('persistedState');
                if (!raw) return;
                const data = JSON.parse(raw);
                if (data?.auth) data.auth.sessionToken = null;
                window.localStorage.setItem('persistedState', JSON.stringify(data));
            } catch {
                // Swallow: a parse error here just means we boot with no persisted state,
                // which is equivalent to the unauthenticated case we're trying to set up.
            }
        });
    });

    test('login form has no detectable WCAG A/AA violations', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/login');

        // Wait for login form to load
        await expect(page.getByRole('textbox', { name: 'Email Address' })).toBeVisible();

        const results = await makeAxeBuilder().analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
