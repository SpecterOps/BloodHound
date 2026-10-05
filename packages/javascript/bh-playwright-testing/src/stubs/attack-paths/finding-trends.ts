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

// Deterministic environment the "With graph data" graph spec navigates to and asserts on. The id
// matches the owner_objectid baked into the meta-node/meta-tree fixtures below so the graph resolves
// a root node, and `name` is the label RootMetaNodeAnnotation renders once the graph settles.
export const environment = {
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

// Tier Zero asset group tag. useSelectedTag / useHighestPrivilegeTagId / useContextSeverity read this
// (matched by exposures[].asset_group_tag.id) to label and scope the root meta-node annotation.
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

// Root Meta node (Tier Zero) keyed by node id; getLatestMetaNode returns this and its first key
// becomes the graph's rootMetaNodeId.
const metaNodes = {
    '1564342163': {
        color: '#000',
        data: {
            isTierZero: true,
            kinds: ['Meta', 'Tag_Tier_Zero'],
            level: 0,
            nodetype: 'Meta',
            owner_objectid: environment.id,
            zone: 'Tag_Tier_Zero',
        },
        border: { color: 'black' },
        image: '/ui/meta.png',
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'NO NAME' },
        size: 1,
    },
};

// The full meta tree: the root node plus child meta-nodes and the relationships linking them to the
// root. getAttackPathsGraph returns this; merged with metaNodes it forms the rendered graph.
const metaTrees = {
    ...metaNodes,
    '1564342165': {
        color: '#000',
        data: {
            isTierZero: false,
            kinds: ['Meta'],
            nodetype: 'Meta',
            owner_objectid: environment.id,
            zone: 'Tag_Tier_Zero',
        },
        border: { color: 'black' },
        image: '/ui/meta.png',
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'NO NAME' },
        size: 1,
    },
    '1564342167': {
        color: '#000',
        data: {
            isTierZero: false,
            kinds: ['Meta'],
            nodetype: 'Meta',
            owner_objectid: environment.id,
            zone: 'Tag_Tier_Zero',
        },
        border: { color: 'black' },
        image: '/ui/meta.png',
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'NO NAME' },
        size: 1,
    },
    rel_19954328980: {
        color: '3a5464',
        data: { composite_risk_impact_count: 9, composite_risk_impact_percent: 60 },
        id: 19954328980,
        end2: { arrow: true },
        id1: '1564342165',
        id2: '1564342163',
        label: { text: 'GenericAll' },
    },
    rel_19954328982: {
        color: '3a5464',
        data: { composite_risk_impact_count: 4, composite_risk_impact_percent: 26.67 },
        id: 19954328982,
        end2: { arrow: true },
        id1: '1564342167',
        id2: '1564342163',
        label: { text: 'MemberOf' },
    },
};

// The two findings surfaced in the right information panel. They line up with the GenericAll /
// MemberOf edges in metaTrees so the graph and findings list tell a consistent story.
const findingNames = ['T0GenericAll', 'T0MemberOf'];

// getPostureFindingTrends (finding-trends) data, keyed later by finding name. Header renders
// display_title and composite_risk from these entries.
const findingTrends = [
    {
        environment_ids: [environment.id],
        composite_risk: 60,
        finding: 'T0GenericAll',
        impact_count: 246,
        exposure_count: 24,
        finding_count_start: 14,
        finding_count_end: 17,
        finding_count_increase: 12,
        finding_count_decrease: 9,
        archived_count: 22,
        display_title: 'GenericAll Privileges on Objects in Privilege Zone',
        display_type: 'Tier Zero Attack Paths',
    },
    {
        environment_ids: [environment.id],
        composite_risk: 26.666666666666668,
        finding: 'T0MemberOf',
        impact_count: 84,
        exposure_count: 6,
        finding_count_start: 10,
        finding_count_end: 4,
        finding_count_increase: 20,
        finding_count_decrease: 26,
        archived_count: 45,
        display_title: 'Non-Certified Principal with Privileges in Privilege Zone',
        display_type: 'Tier Zero Attack Paths',
    },
];

// One sparkline point per finding. mapAvailableFindingSparklines keys these by data[0].Finding and
// FindingsInformationPanel treats a finding with no points as "unavailable", so each needs >=1 point.
const findingSparkline = (finding: string, compositeRisk: number) => [
    {
        id: 1,
        DomainSID: environment.id,
        Finding: finding,
        CompositeRisk: compositeRisk,
        FindingCount: 1,
        ImpactedAssetCount: 1,
        created_at: '2026-08-24T21:42:00Z',
        updated_at: '2026-08-24T21:42:00Z',
    },
];

// One relationship-detail row per finding. mapAvailableFindingDetails reads Finding /
// ComboGraphRelationID to build findingNameCounts and findingNameToEdgeMap (rel_<id>).
const findingDetail = (finding: string, comboGraphRelationId: number) => ({
    data: [{ Finding: finding, ComboGraphRelationID: comboGraphRelationId }],
    count: 1,
});

// getFindings (findings/{name}) content, keyed by finding name. Header renders `title` and
// FindingInformation renders the description / remediation markdown. Trimmed from the full API
// payload to the fields the components read while keeping enough prose to exercise MarkdownContent.
const findingContent: Record<string, Record<string, string>> = {
    T0GenericAll: {
        title: 'GenericAll Privileges on Objects in Privilege Zone',
        type: 'Privilege Zone Attack Paths',
        short_description:
            'The "GenericAll" privilege grants principals the ability to perform nearly any action against the object, including abusable actions such as resetting user passwords and adding new users to security groups. "GenericAll" is also known as "Full Control".',
        long_description:
            'The "GenericAll" privilege grants principals the ability to perform nearly any action against the object, including abusable actions such as resetting user passwords and adding new users to security groups.\n\nOnly those users belonging to Tier Zero domain groups should have control over Tier Zero users, groups, computers, and special objects.',
        short_remediation:
            'Remove the "GenericAll" privilege that the Principal outside the Privilege Zone holds against the object in the Privilege Zone.',
        long_remediation:
            'Deny all principals in the domain the ability to control a Tier Zero asset except where that principal is itself also a Tier Zero asset.',
        references:
            '### MITRE ATT&CK\n* [ATT&CK T1098: Account Manipulation](https://attack.mitre.org/techniques/T1098/)',
    },
    T0MemberOf: {
        title: 'Non-Certified Principal with Privileges in Privilege Zone',
        type: 'Privilege Zone Attack Paths',
        short_description:
            'Group members in Active Directory retain all privileges held by the group, therefore, if a group is in a Privilege Zone, then the members of that group must also be part of that Privilege Zone.',
        long_description:
            'Group members in Active Directory retain all privileges held by the group, therefore, if a group is in a Privilege Zone, then the members of that group must also be part of that Privilege Zone.',
        short_remediation:
            'Either certify the group member from the Certification page, or remove the group member from the Privilege Zone group.',
        long_remediation:
            'If the object is an expected member, certify the object in the certifications page. Otherwise, remove the object from the group it belongs to.',
        references:
            '### How Attackers Abuse This Attack Path\n* [BloodHound Docs: MemberOf](https://bloodhound.specterops.io/resources/edges/member-of)',
    },
};

/**
 * Stubs the Attack Paths graph endpoints so the "With graph data" spec renders a deterministic,
 * hermetic Tier Zero graph for `environment`:
 *   - `available-domains` returns the single environment (also keeps the "No Data" dialog closed).
 *   - `asset-group-tags` returns the Tier Zero tag used to label/scope the root node annotation.
 *   - `meta-nodes` / `meta-trees` return the fixture graph so regraph settles and the environment
 *     name label appears.
 *   - `available-types` / `sparkline` / `details` / `finding-trends` return the two fixture findings
 *     so the right information panel lists findings instead of "No risks identified".
 *   - `findings/{name}` returns each finding's title / description / remediation content so the
 *     finding header renders its title and the expanded panel renders its markdown.
 *   - `custom-nodes` returns an empty array so the global saga resolves without escaping to the
 *     real backend.
 * Only GET traffic is handled; anything else falls through to lower-priority handlers.
 */
export async function installAttackPathsStubs(page: Page): Promise<void> {
    await page.route('**/api/v2/available-domains', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: [environment] } });
    });

    await page.route('**/api/v2/asset-group-tags?*', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: { tags: [tierZeroTag] } } });
    });

    await page.route(`**/api/v2/meta-nodes/${environment.id}*`, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: metaNodes } });
    });

    await page.route(`**/api/v2/meta-trees/${environment.id}*`, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: metaTrees } });
    });

    await page.route('**/api/v2/custom-nodes', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: [] } });
    });

    await page.route(`**/api/v2/domains/${environment.id}/available-types*`, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: findingNames } });
    });

    await page.route(`**/api/v2/domains/${environment.id}/sparkline*`, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        const finding = new URL(route.request().url()).searchParams.get('finding') ?? '';
        const compositeRisk = findingTrends.find((trend) => trend.finding === finding)?.composite_risk ?? 0;
        return route.fulfill({ json: { data: findingSparkline(finding, compositeRisk), count: 1 } });
    });

    await page.route(`**/api/v2/domains/${environment.id}/details*`, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        const finding = new URL(route.request().url()).searchParams.get('finding') ?? '';
        const comboGraphRelationId = finding === 'T0GenericAll' ? 19954328980 : 19954328982;
        return route.fulfill({ json: findingDetail(finding, comboGraphRelationId) });
    });

    await page.route('**/api/v2/findings/*', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        const finding = new URL(route.request().url()).pathname.split('/').pop() ?? '';
        return route.fulfill({ json: { data: findingContent[finding] ?? {} } });
    });

    await page.route('**/api/v2/attack-paths/finding-trends*', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({
            json: {
                data: {
                    findings: findingTrends,
                    total_finding_count_start: 24,
                    total_finding_count_end: 21,
                },
                start: null,
                end: null,
                environments: [environment.id],
            },
        });
    });
}
