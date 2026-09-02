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
import { hideBySelector } from 'bh-playwright-testing/axe';
import { expectNoAccessibilityViolations, test } from '../../fixtures';

const administrator = {
    sso_provider_id: null,
    AuthSecret: {
        digest_method: 'argon2',
        expires_at: '2027-01-01T12:00:00Z',
        totp_activated: false,
        id: 31,
        created_at: '2026-01-01T12:00:00Z',
        updated_at: '2026-01-01T12:00:00Z',
        deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
    },
    roles: [
        {
            name: 'Administrator',
            description: 'Administrator',
            permissions: [{ authority: 'auth', name: 'ManageUsers' }],
            id: 4,
            created_at: '2026-01-01T12:00:00Z',
            updated_at: '2026-01-01T12:00:00Z',
            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
        },
    ],
    first_name: 'Test',
    last_name: 'Administrator',
    email_address: 'test-admin@example.com',
    principal_name: 'test_admin',
    last_login: '0001-01-01T00:00:00Z',
    is_disabled: false,
    eula_accepted: true,
    id: 'user-1',
    created_at: '2026-01-01T12:00:00Z',
    updated_at: '2026-01-01T12:00:00Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
};

const token = {
    name: 'automation token',
    hmac_method: 'hmac-sha2-256',
    last_access: '2026-01-02T12:00:00Z',
    id: 'token-1',
    created_at: '2026-01-01T12:00:00Z',
    updated_at: '2026-01-01T12:00:00Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
    expires_at: null,
};

const openPage = async (page: Page) => {
    await page.goto('/ui/administration/manage-users');
    await page.getByRole('button', { name: 'Toggle Navigation' }).click();
};

const openUserActions = async (page: Page) => {
    await page
        .getByRole('row', { name: /test_admin/ })
        .getByRole('button', { name: 'Show user actions' })
        .click();
};

const openTokenManagement = async (page: Page) => {
    await openUserActions(page);
    await page.getByRole('menuitem', { name: 'Generate / Revoke API Tokens' }).click();
};

test.describe('Administration - Manage Users - has no detectable WCAG A/AA violations', () => {
    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v2/self', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: administrator } });
        });

        await page.route('**/api/v2/config**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({
                json: {
                    data: [
                        { key: 'auth.api_tokens', value: { enabled: true } },
                        { key: 'auth.api_token_expiration', value: { enabled: false, expiration_period: '90' } },
                    ],
                },
            });
        });

        await page.route('**/api/v2/sso-providers', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: [] } });
        });

        await page.route('**/api/v2/roles**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: { roles: administrator.roles } } });
        });

        await page.route('**/api/v2/available-domains', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: [] } });
        });

        await page.route('**/api/v2/bloodhound-users**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }
            const users = [administrator];
            const pathname = new URL(route.request().url()).pathname;
            if (pathname === '/api/v2/bloodhound-users') {
                return route.fulfill({ json: { data: { users } } });
            }

            const user = users.find(({ id }) => pathname === `/api/v2/bloodhound-users/${id}`);
            if (user) {
                return route.fulfill({ json: { data: user } });
            }

            await route.fallback();
        });
    });

    test('empty page', async ({ page, makeAxeBuilder }, testInfo) => {
        // Override the users call to be empty users list
        await page.route('**/api/v2/bloodhound-users**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }
            const users: (typeof administrator)[] = [];
            const pathname = new URL(route.request().url()).pathname;
            if (pathname === '/api/v2/bloodhound-users') {
                return route.fulfill({ json: { data: { users } } });
            }

            const user = users.find(({ id }) => pathname === `/api/v2/bloodhound-users/${id}`);
            if (user) {
                return route.fulfill({ json: { data: user } });
            }

            await route.fallback();
        });

        await openPage(page);
        await page.getByRole('heading', { name: 'Manage Users' }).waitFor({ state: 'visible' });
        await page.getByRole('columnheader', { name: 'Username' }).waitFor({ state: 'visible' });
        await page.getByRole('row').nth(1).waitFor({ state: 'hidden' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('page with users', async ({ page, makeAxeBuilder }, testInfo) => {
        await openPage(page);

        await page.getByText('test_admin', { exact: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('user actions menu displayed', async ({ page, makeAxeBuilder }, testInfo) => {
        await openPage(page);
        await openUserActions(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('menuitem', { name: 'Update User' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="menu"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('update user dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await openPage(page);
        await openUserActions(page);
        await page.getByRole('menuitem', { name: 'Update User' }).click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Update User Dialog' }).waitFor({ state: 'visible' });
        await page.getByLabel('Email Address').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[data-testid="update-user-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('change password dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await openPage(page);
        await openUserActions(page);
        await page.getByRole('menuitem', { name: 'Change Password' }).click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Change Password' }).waitFor({ state: 'visible' });

        await page.getByRole('textbox', { name: 'New Password', exact: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[data-testid="password-dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('generate/revoke API tokens dialog - no tokens', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/tokens**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: { tokens: [] } } });
        });

        await openPage(page);
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Generate/Revoke API Tokens' }).waitFor({ state: 'visible' });
        await page.getByText('No tokens available').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[data-testid="user-token-management-dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('create token dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/tokens**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: { tokens: [] } } });
        });

        await openPage(page);
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('button', { name: 'Create Token' }).click();
        await page.getByRole('dialog', { name: 'Create User Token' }).waitFor({ state: 'visible' });

        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        const results = await makeAxeBuilder().include('[data-testid="create-user-token-dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('auth token dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/tokens**', async (route) => {
            const method = route.request().method();

            if (method === 'GET') {
                return route.fulfill({ json: { data: { tokens: [] } } });
            }

            if (method === 'POST' && new URL(route.request().url()).pathname === '/api/v2/tokens') {
                return route.fulfill({ json: { data: { ...token, key: 'generated-auth-token-key' } } });
            }

            await route.fallback();
        });

        await openPage(page);
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('button', { name: 'Create Token' }).click();
        await page.getByRole('dialog', { name: 'Create User Token' }).waitFor({ state: 'visible' });

        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        await page.getByLabel('Token Name').fill(token.name);
        await page.getByRole('dialog', { name: 'Create User Token' }).getByRole('button', { name: 'Save' }).click();

        await page.getByRole('dialog', { name: 'Auth Token' }).waitFor({ state: 'visible' });
        await page.getByText('Key: generated-auth-token-key').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[data-testid="user-token-dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('generate/revoke API tokens dialog - with token', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/tokens**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: { tokens: [token] } } });
        });

        await openPage(page);
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('row', { name: /automation token/ }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[data-testid="user-token-management-dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('revoke token dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/tokens**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: { tokens: [token] } } });
        });

        await openPage(page);
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page
            .getByRole('row', { name: /automation token/ })
            .getByRole('button', { name: 'Revoke' })
            .click();

        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        await page.getByRole('dialog', { name: 'Revoke "automation token" Auth Token' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[data-testid="token-revoke-dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('create user form', async ({ page, makeAxeBuilder }, testInfo) => {
        await openPage(page);
        await page.getByRole('button', { name: 'Create User' }).click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Create User' }).waitFor({ state: 'visible' });
        await page.getByLabel('Email Address').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[data-testid="create-user-dialog_form"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
