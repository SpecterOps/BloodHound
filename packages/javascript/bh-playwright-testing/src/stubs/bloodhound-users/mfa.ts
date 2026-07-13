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

const DEFAULT_MFA_QR_CODE =
    'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=';
const DEFAULT_MFA_TOTP_SECRET = 'ABCDEFGHIJKLMNOP';

export type MFAEnrollmentStubOptions = {
    qrCode?: string;
    totpSecret?: string;
};

/**
 * Stubs the happy path for enabling MFA without changing the real user's MFA state.
 * Install before clicking the MFA toggle so the dialog can advance through every step.
 */
export async function installMFAEnrollmentStub(page: Page, opts: MFAEnrollmentStubOptions = {}): Promise<void> {
    const qrCode = opts.qrCode ?? DEFAULT_MFA_QR_CODE;
    const totpSecret = opts.totpSecret ?? DEFAULT_MFA_TOTP_SECRET;

    await page.route(/\/api\/v2\/bloodhound-users\/[^/]+\/mfa$/, async (route) => {
        if (route.request().method() !== 'POST') {
            return route.fallback();
        }

        return route.fulfill({
            json: {
                data: {
                    qr_code: qrCode,
                    totp_secret: totpSecret,
                },
            },
        });
    });

    await page.route(/\/api\/v2\/bloodhound-users\/[^/]+\/mfa-activation$/, async (route) => {
        if (route.request().method() !== 'POST') {
            return route.fallback();
        }

        return route.fulfill({ json: { data: {} } });
    });
}
