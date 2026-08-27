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

type CollectorAsset = {
    name: string;
    download_url: string;
    checksum_download_url: string;
    os: string;
    arch: string;
};

type CollectorManifest = {
    version: string;
    version_meta: { major: number; minor: number; patch: number; prerelease: string };
    release_date: string;
    release_assets: CollectorAsset[];
};

type CommunityCollectorType = 'sharphound' | 'azurehound';
type EnterpriseCollectorType = 'sharphound_enterprise' | 'azurehound_enterprise' | 'openhound';

type CommunityManifest = Record<CommunityCollectorType, CollectorManifest[]>;
type EnterpriseManifest = Record<EnterpriseCollectorType, CollectorManifest[]>;

const GITHUB = 'https://github.com/SpecterOps';
const KENNEL_DOWNLOAD = 'https://test.bloodhoundenterprise.io/api/v2/kennel/download';

const asset = (name: string, base: string, os: string, arch: string): CollectorAsset => ({
    name,
    download_url: `${base}/${name}`,
    checksum_download_url: `${base}/${name}.sha256`,
    os,
    arch,
});

const release = (
    version: string,
    prerelease: string,
    release_date: string,
    release_assets: CollectorAsset[]
): CollectorManifest => {
    const [major, minor, patch] = version.replace(/^v/, '').split('-')[0].split('.').map(Number);
    return { version, version_meta: { major, minor, patch, prerelease }, release_date, release_assets };
};

const COMMUNITY_MANIFEST: CommunityManifest = {
    sharphound: [
        release('v2.14.0', '', '2026-07-24T15:24:43Z', [
            asset(
                'SharpHound_v2.14.0_windows_x86.zip',
                `${GITHUB}/SharpHound/releases/download/v2.14.0`,
                'windows',
                'x86'
            ),
        ]),
        release('v2.14.0-rc1', 'rc1', '2026-07-21T14:11:43Z', [
            asset(
                'SharpHound_v2.14.0-rc1_windows_x86.zip',
                `${GITHUB}/SharpHound/releases/download/v2.14.0-rc1`,
                'windows',
                'x86'
            ),
        ]),
    ],
    azurehound: [
        release('v3.1.0', '', '2026-08-17T18:25:34Z', [
            asset(
                'AzureHound_v3.1.0_windows_amd64.zip',
                `${GITHUB}/AzureHound/releases/download/v3.1.0`,
                'windows',
                'amd64'
            ),
            asset(
                'AzureHound_v3.1.0_darwin_arm64.zip',
                `${GITHUB}/AzureHound/releases/download/v3.1.0`,
                'darwin',
                'arm64'
            ),
            asset(
                'AzureHound_v3.1.0_linux_amd64.zip',
                `${GITHUB}/AzureHound/releases/download/v3.1.0`,
                'linux',
                'amd64'
            ),
        ]),
    ],
};

const ENTERPRISE_MANIFEST: EnterpriseManifest = {
    sharphound_enterprise: [
        release('v2.15.0', '', '2026-08-17T18:33:34Z', [
            asset('SharpHoundEnterpriseService_v2.15.0_windows_x86.zip', KENNEL_DOWNLOAD, 'windows', 'x86'),
        ]),
        release('v2.15.0-rc1', 'rc1', '2026-08-11T15:13:29Z', [
            asset('SharpHoundEnterpriseService_v2.15.0-rc1_windows_x86.zip', KENNEL_DOWNLOAD, 'windows', 'x86'),
        ]),
    ],
    azurehound_enterprise: [
        release('v3.1.0', '', '2026-08-17T18:34:51Z', [
            asset('AzureHoundEnterprise_v3.1.0_windows_amd64.zip', KENNEL_DOWNLOAD, 'windows', 'amd64'),
            asset('AzureHoundEnterprise_v3.1.0_windows_arm64.zip', KENNEL_DOWNLOAD, 'windows', 'arm64'),
        ]),
    ],
    openhound: [],
};

/**
 * Stubs the community collector manifest endpoint (`GET /api/v2/kennel/manifest`) so the
 * Download Collectors page renders SharpHound and AzureHound cards without depending on the
 * live GitHub-backed manifest. Install before navigation. Non-GET traffic falls through to
 * any lower-priority route handlers.
 */
export async function installCommunityCollectorsStub(
    page: Page,
    manifest: CommunityManifest = COMMUNITY_MANIFEST
): Promise<void> {
    await page.route('**/api/v2/kennel/manifest', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();

        return route.fulfill({ json: { data: manifest } });
    });
}

/**
 * Stubs the enterprise collector manifest endpoint (`GET /api/v2/kennel/enterprise-manifest`) so
 * the Download Collectors page renders SharpHound/AzureHound Enterprise cards without depending on
 * the live S3-backed manifest. Install before navigation. Non-GET traffic falls through to any
 * lower-priority route handlers.
 */
export async function installEnterpriseCollectorsStub(
    page: Page,
    manifest: EnterpriseManifest = ENTERPRISE_MANIFEST
): Promise<void> {
    await page.route('**/api/v2/kennel/enterprise-manifest', async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();

        return route.fulfill({ json: { data: manifest } });
    });
}

/**
 * Convenience helper that installs both the community and enterprise collector manifest stubs so
 * the Download Collectors page fully renders with data. Install before navigation.
 */
export async function installCollectorsStub(page: Page): Promise<void> {
    await installCommunityCollectorsStub(page);
    await installEnterpriseCollectorsStub(page);
}
