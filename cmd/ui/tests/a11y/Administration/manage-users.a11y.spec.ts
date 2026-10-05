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
import { test } from 'bh-playwright-testing';
import { hideBySelector } from 'bh-playwright-testing/axe';

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

    test('empty page', async ({ page, goAndWaitFor, checkA11y }) => {
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

        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await page.getByRole('columnheader', { name: 'Username' }).waitFor();
        await page.getByRole('row').nth(1).waitFor({ state: 'hidden' });

        await checkA11y();
    });

    test('page with users', async ({ page, goAndWaitFor, checkA11y }) => {
        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));

        await page.getByText('test_admin', { exact: true }).waitFor();

        await checkA11y();
    });

    test('user actions menu displayed', async ({ page, goAndWaitFor, checkA11y }) => {
        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await openUserActions(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('menuitem', { name: 'Update User' }).waitFor();

        await checkA11y({ include: '[role="menu"]' });
    });

    test('update user dialog', async ({ page, goAndWaitFor, checkA11y }) => {
        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await openUserActions(page);
        await page.getByRole('menuitem', { name: 'Update User' }).click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Update User Dialog' }).waitFor();
        await page.getByLabel('Email Address').waitFor();

        await checkA11y({ include: '[data-testid="update-user-dialog"]' });
    });

    test('change password dialog', async ({ page, goAndWaitFor, checkA11y }) => {
        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await openUserActions(page);
        await page.getByRole('menuitem', { name: 'Change Password' }).click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Change Password' }).waitFor();
        await page.getByRole('textbox', { name: 'New Password Confirmation' }).waitFor();
        await page.getByRole('textbox', { name: 'New Password', exact: true }).waitFor({ state: 'visible' });

        await checkA11y({ include: '[data-testid="password-dialog"]' });
    });

    test('generate/revoke API tokens dialog - no tokens', async ({ page, goAndWaitFor, checkA11y }) => {
        await page.route('**/api/v2/tokens**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: { tokens: [] } } });
        });

        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Generate/Revoke API Tokens' }).waitFor();
        await page.getByText('No tokens available').waitFor();

        await checkA11y({ include: '[data-testid="user-token-management-dialog"]' });
    });

    test('create token dialog', async ({ page, goAndWaitFor, checkA11y }) => {
        await page.route('**/api/v2/tokens**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: { tokens: [] } } });
        });

        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('button', { name: 'Create Token' }).click();
        await page.getByRole('dialog', { name: 'Create User Token' }).waitFor();

        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        await checkA11y({ include: '[data-testid="create-user-token-dialog"]' });
    });

    test('auth token dialog', async ({ page, goAndWaitFor, checkA11y }) => {
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

        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('button', { name: 'Create Token' }).click();
        await page.getByRole('dialog', { name: 'Create User Token' }).waitFor();

        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        await page.getByLabel('Token Name').fill(token.name);
        await page.getByRole('dialog', { name: 'Create User Token' }).getByRole('button', { name: 'Save' }).click();

        await page.getByRole('dialog', { name: 'Auth Token' }).waitFor();
        await page.getByText('Key: generated-auth-token-key').waitFor();

        await checkA11y({ include: '[data-testid="user-token-dialog"]' });
    });

    test('generate/revoke API tokens dialog - with token', async ({ page, goAndWaitFor, checkA11y }) => {
        await page.route('**/api/v2/tokens**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: { tokens: [token] } } });
        });

        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('row', { name: /automation token/ }).waitFor();

        await checkA11y({ include: '[data-testid="user-token-management-dialog"]' });
    });

    test('revoke token dialog', async ({ page, goAndWaitFor, checkA11y }) => {
        await page.route('**/api/v2/tokens**', async (route) => {
            if (route.request().method() !== 'GET') return route.fallback();
            await route.fulfill({ json: { data: { tokens: [token] } } });
        });

        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await openTokenManagement(page);

        await hideBySelector(page, '#content-wrapper');

        await page
            .getByRole('row', { name: /automation token/ })
            .getByRole('button', { name: 'Revoke' })
            .click();

        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        await page.getByRole('dialog', { name: 'Revoke "automation token" Auth Token' }).waitFor();

        await checkA11y({ include: '[data-testid="token-revoke-dialog"]' });
    });

    test('create user form', async ({ page, goAndWaitFor, checkA11y }) => {
        await goAndWaitFor('/ui/administration/manage-users', page.getByRole('heading', { name: 'Manage Users' }));
        await page.getByRole('button', { name: 'Create User' }).click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Create User' }).waitFor();
        await page.getByLabel('Email Address').waitFor();

        await checkA11y({ include: '[data-testid="create-user-dialog_form"]' });
    });
});
