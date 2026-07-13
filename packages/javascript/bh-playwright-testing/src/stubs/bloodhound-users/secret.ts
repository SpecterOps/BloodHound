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

/**
 * Stubs the update password endpoint so the Reset Password dialog can complete the password change
 * flow without actually updating the real user's password. Install before submitting the password
 * form. Non-PUT traffic falls through to any lower-priority route handlers.
 */
export async function installResetPasswordStub(page: Page): Promise<void> {
    await page.route(/\/api\/v2\/bloodhound-users\/[^/]+\/secret$/, async (route) => {
        if (route.request().method() !== 'PUT') {
            return route.fallback();
        }

        return route.fulfill({ status: 200 });
    });
}
