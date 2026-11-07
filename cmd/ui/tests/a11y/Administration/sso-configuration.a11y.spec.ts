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

import { hideBySelector } from 'bh-playwright-testing/axe';
import { test } from 'bh-playwright-testing';
const authenticatedUser = {
    AuthSecret: {
        expires_at: '9999-01-01T00:00:00Z',
    },
    roles: [{ permissions: [{ authority: 'auth', name: 'ManageProviders' }] }],
    first_name: 'Test',
    last_name: 'Administrator',
    email_address: 'test-admin@example.com',
    principal_name: 'test_admin',
    is_disabled: false,
    eula_accepted: true,
    id: 'user-1',
};

const roles = [
    { id: 1, name: 'Read-Only' },
    { id: 2, name: 'Power User' },
    { id: 3, name: 'Administrator' },
    { id: 4, name: 'Upload-Only' },
];

const samlProvider = {
    id: 1,
    type: 'SAML',
    slug: 'test-idp-1',
    name: 'Test IDP 1',
    details: {
        idp_issuer_uri: 'http://test-idp-1:8081/metadata',
        idp_sso_uri: 'http://test-idp-1.localhost/sso',
        principal_attribute_mappings: null,
        sp_issuer_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1',
        sp_sso_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1/sso',
        sp_metadata_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1/metadata',
        sp_acs_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1/acs',
    },
    login_uri: '',
    callback_uri: '',
    created_at: '2022-02-24T23:38:41.420271Z',
    updated_at: '2022-02-24T23:38:41.420271Z',
    config: {
        auto_provision: { enabled: false, role_provision: false, default_role_id: 1 },
    },
};

test.describe('Administration - SSO Configuration - has no detectable WCAG A/AA violations', () => {
    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v2/self', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: authenticatedUser } });
        });
        await page.route('**/api/v2/roles', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: { roles } } });
        });
        await page.route('**/api/v2/features', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            id: 1,
                            key: 'oidc_support',
                            name: 'OIDC Support',
                            description: 'Enables OIDC identity providers.',
                            enabled: true,
                            user_updatable: false,
                        },
                    ],
                },
            });
        });

        await page.route('**/api/v2/sso-providers', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [] } });
        });
    });

    test('page with no providers', async ({ page, goAndWaitFor, checkA11y }) => {
        await goAndWaitFor(
            '/ui/administration/sso-configuration',
            page.getByRole('heading', { name: 'SSO Configuration' })
        );

        await page.getByText('No SSO Providers found').waitFor({ state: 'visible' });

        await checkA11y();
    });

    test('create providers menu', async ({ page, checkA11y }) => {
        await page.goto('/ui/administration/sso-configuration');
        await page.getByRole('button', { name: 'Create Provider' }).click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('menuitem', { name: 'SAML Provider' }).waitFor({ state: 'visible' });
        await page.getByRole('menuitem', { name: 'OIDC Provider' }).waitFor({ state: 'visible' });

        await checkA11y({ include: '[role="menu"]' });
    });

    test('create SAML provider dialog', async ({ page, checkA11y }) => {
        await page.goto('/ui/administration/sso-configuration');
        await page.getByRole('button', { name: 'Create Provider' }).click();
        await page.getByRole('menuitem', { name: 'SAML Provider' }).click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Create SAML Provider' }).waitFor({ state: 'visible' });
        await page.getByLabel('SAML Provider Name').waitFor({ state: 'visible' });

        await checkA11y({ include: '[data-testid="create-saml-provider-dialog"]' });
    });

    test('create OIDC provider dialog', async ({ page, checkA11y }) => {
        await page.goto('/ui/administration/sso-configuration');
        await page.getByRole('button', { name: 'Create Provider' }).click();
        await page.getByRole('menuitem', { name: 'OIDC Provider' }).click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('dialog', { name: 'Create OIDC Provider' }).waitFor({ state: 'visible' });
        await page.getByLabel('OIDC Provider Name').waitFor({ state: 'visible' });

        await checkA11y({ include: '[data-testid="create-oidc-provider-dialog"]' });
    });

    test('page with providers', async ({ page, goAndWaitFor, checkA11y }) => {
        await page.route('**/api/v2/sso-providers', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [samlProvider] } });
        });

        await goAndWaitFor('/ui/administration/sso-configuration', page.getByRole('button', { name: 'Test IDP 1' }));

        await page.getByRole('cell', { name: 'SAML' }).waitFor({ state: 'visible' });

        await checkA11y();
    });
});
