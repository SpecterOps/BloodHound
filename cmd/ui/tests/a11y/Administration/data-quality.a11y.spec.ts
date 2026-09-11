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

const dataQualityResult = {
    acls: 34,
    aiacas: 1,
    certtemplates: 8,
    computers: 12,
    containers: 5,
    created_at: '2026-08-01T12:00:00Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
    domains: 1,
    enterprisecas: 1,
    gpos: 4,
    groups: 9,
    issuancepolicies: 2,
    local_group_completeness: 0.75,
    ntauthstores: 1,
    ous: 6,
    relationships: 56,
    rootcas: 1,
    session_completeness: 0.5,
    sessions: 23,
    updated_at: '2026-08-01T12:00:00Z',
    users: 10,
};

test.describe('Administration - Data Quality - has no detectable WCAG A/AA violations', () => {
    test('empty page', async ({ page, goAndWaitFor, checkA11y }) => {
        await page.route('**/api/v2/available-domains', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [] } });
        });

        await goAndWaitFor(
            '/ui/administration/data-quality',
            page.getByText('No Domain or Tenant Selected', { exact: true })
        );
        await checkA11y();
    });

    test('page with results', async ({ page, goAndWaitFor, checkA11y }) => {
        await page.route('**/api/v2/available-domains', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            collected: true,
                            exposures: [],
                            hygiene_attack_paths: 0,
                            id: 'example-domain-id',
                            impactValue: 0,
                            name: 'EXAMPLE.COM',
                            type: 'active-directory',
                        },
                    ],
                },
            });
        });
        await page.route('**/api/v2/ad-domains/example-domain-id/data-quality-stats**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    count: 1,
                    data: [dataQualityResult],
                    limit: 1,
                    skip: 0,
                },
            });
        });

        await goAndWaitFor('/ui/administration/data-quality', page.getByText('Group Completeness'));

        await checkA11y();
    });
});
