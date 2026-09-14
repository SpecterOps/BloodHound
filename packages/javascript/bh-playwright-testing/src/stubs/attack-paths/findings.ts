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

// Deterministic environment the table findings belong to. Its id/name match the environment_id and
// environment_name baked into the findings below so the Environment column and Environment filter tell
// a consistent story, and a single collected environment keeps the "No Data" upload dialog closed.
const environment = {
    type: 'active-directory',
    name: 'SEVENKINGDOMS.LOCAL',
    id: 'S-1-5-21-2768881856-185705006-2548489946',
    collected: true,
    impactValue: 74,
    hygiene_attack_paths: 0,
    exposures: [
        {
            exposure_percent: 74,
            asset_group_tag: { id: 1, type: 1, name: 'Tier Zero', position: 1 },
        },
    ],
};

// Tier Zero asset group tag. The Zone column / Zone filter read this (matched by asset_group_tag_id)
// to label the findings' zone.
const tierZeroTag = {
    id: 1,
    type: 1,
    kind_id: 173,
    name: 'Tier Zero',
    description: 'Tier Zero',
    created_at: '2025-04-15T21:02:26.504736Z',
    created_by: 'BloodHound',
    updated_at: '2026-06-09T17:01:52.688496Z',
    updated_by: 'playwright',
    deleted_at: null,
    deleted_by: null,
    position: 1,
    require_certify: true,
    analysis_enabled: true,
    glyph: 'gem',
};

// One template per distinct attack-path finding captured in table.json, paired with how many rows of
// that finding the HAR returned. Expanded below into the 26-row page the findings table renders.
const findingTemplates: { template: Record<string, unknown>; rows: number }[] = [
    {
        rows: 17,
        template: {
            severity: 'moderate',
            finding: 'T0GenericAll',
            title: 'GenericAll Privileges on Objects in Privilege Zone',
            finding_type: 'relationship',
            status: 'active',
        },
    },
    {
        rows: 4,
        template: {
            severity: 'high',
            finding: 'T0MarkSensitive',
            title: 'Tier Zero Objects Lack Kerberos Delegation Protection',
            finding_type: 'list',
            status: 'active',
        },
    },
    {
        rows: 4,
        template: {
            severity: 'low',
            finding: 'T0MemberOf',
            title: 'Non-Certified Principal with Privileges in Privilege Zone',
            finding_type: 'relationship',
            status: 'accepted',
        },
    },
    {
        rows: 1,
        template: {
            severity: 'low',
            finding: 'T0GPLink',
            title: 'Non-Certified GPO Linked to Privilege Zone OU',
            finding_type: 'relationship',
            status: 'active',
        },
    },
];

// Expand the templates into the flat 26-row page. Each row gets a deterministic target principal so the
// Target Principal column renders distinct, stable values without embedding every captured row verbatim.
const findings = findingTemplates.flatMap(({ template, rows }, groupIndex) =>
    Array.from({ length: rows }, (_unused, rowIndex) => ({
        ...template,
        platform: 'Active Directory',
        environment_id: environment.id,
        environment_name: environment.name,
        asset_group_tag_id: tierZeroTag.id,
        zone_name: tierZeroTag.name,
        source_principal_id: '',
        source_principal_kind: '',
        source_principal_name: '',
        target_principal_id: `${environment.id}-${1000 + groupIndex * 100 + rowIndex}`,
        target_principal_kind: 'User',
        target_principal_name: `PRINCIPAL-${groupIndex}-${rowIndex}@${environment.name}`,
        first_seen: '2026-06-05T05:09:50.085107Z',
        last_seen: '2026-08-25T03:42:11.149145Z',
    }))
);

/**
 * Stubs the Attack Paths table endpoints so the "With attack paths" spec renders a deterministic,
 * hermetic findings table:
 *   - `available-domains` returns the single environment the findings belong to (also keeps the
 *     "No Data" dialog closed) and populates the Environment filter.
 *   - `asset-group-tags` returns the Tier Zero tag so the Zone column / filter resolve.
 *   - `attack-paths/findings` returns the 26-row findings page (skip/limit/count) that
 *     `useUnifiedFindings` reads, so the table body and the "26 Findings" count render instead of the
 *     loading or empty state.
 * Only GET traffic is handled; anything else falls through to lower-priority handlers.
 */
export async function installAttackPathsTableStubs(page: Page): Promise<void> {
    await page.route('**/api/v2/available-domains', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: [environment] } });
    });

    await page.route('**/api/v2/asset-group-tags?*', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: { tags: [tierZeroTag] } } });
    });

    await page.route('**/api/v2/attack-paths/findings?*', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({
            json: { data: findings, skip: 0, limit: 100, count: findings.length },
        });
    });
}
