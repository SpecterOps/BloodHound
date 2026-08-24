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

import { test } from 'bh-playwright-testing';
import { hideBySelector } from 'bh-playwright-testing/axe';
const earlyAccessFeature = {
    id: 1,
    name: 'Example Early Access Feature',
    key: 'example_early_access_feature',
    description: 'Enables an example feature available for early access testing.',
    enabled: false,
    user_updatable: true,
};

test.describe('Administration - Early Access Features - has no detectable WCAG A/AA violations', () => {
    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v2/self', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: {
                        AuthSecret: {
                            expires_at: '9999-01-01T00:00:00Z',
                        },
                        roles: [{ permissions: [{ authority: 'app', name: 'WriteAppConfig' }] }],
                        first_name: 'Test',
                        last_name: 'Administrator',
                        email_address: 'test-admin@example.com',
                        principal_name: 'test_admin',
                        is_disabled: false,
                        eula_accepted: true,
                        id: 'user-1',
                    },
                },
            });
        });
        await page.route('**/api/v2/features', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [earlyAccessFeature] } });
        });
        await page.route('**/api/v2/config', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [] } });
        });
    });

    test('notice dialog', async ({ checkA11y, goAndWaitFor, page }) => {
        await goAndWaitFor(
            '/ui/administration/early-access-features',
            page.getByRole('heading', { name: 'Heads up!' })
        );

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('heading', { name: 'Heads up!' }).waitFor();
        await page.getByRole('button', { name: 'Take me back' }).waitFor();
        await page.getByRole('button', { name: 'I understand, show me the new stuff!' }).waitFor();

        await checkA11y({ include: '[data-testid="early-access-features-warning-dialog"]' });
    });

    test('page', async ({ checkA11y, goAndWaitFor, page }) => {
        await goAndWaitFor(
            '/ui/administration/early-access-features',
            page.getByRole('heading', { name: 'Heads up!' })
        );

        await page.getByRole('button', { name: 'I understand, show me the new stuff!' }).click();

        await page.getByRole('heading', { name: 'Heads up!' }).waitFor({ state: 'hidden' });
        await page.getByRole('heading', { name: 'Early Access Features' }).waitFor();
        await page.getByText(earlyAccessFeature.description).waitFor();

        await checkA11y();
    });
});
