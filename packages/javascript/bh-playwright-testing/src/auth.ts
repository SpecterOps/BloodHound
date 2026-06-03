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
import { expect } from '@playwright/test';
import type { Theme } from './themes';

export type LoginAndSnapshotThemesOptions = {
    page: Page;
    username: string;
    password: string;
    // Resolves a per-theme `storageState` file path. Callers typically wrap
    // `authStorageStateFor(theme)` from `./themes` with `path.resolve(...)`.
    storageStatePathFor: (theme: Theme) => string;
    // Optional hook for dismissing any post-login overlay (e.g. an upload dialog) that
    // could intercept the dark-mode toggle click. Invoked after the login form clears.
    dismissPostLogin?: (page: Page) => Promise<void>;
};

// Login once and snapshot `storageState` for both light and dark themes.
//
// Capturing both snapshots from a single session avoids the parallel-login race where two
// setups as the same user would invalidate each other's session. This relies on the BH
// shared UI shell (login form labels "Email Address" / "Password", the LOGIN submit button,
// the `global_nav-dark-mode` toggle, and the `persistedState` localStorage key written by
// the global store's throttled subscriber) being identical across consumers.
export async function loginAndSnapshotThemes(opts: LoginAndSnapshotThemesOptions): Promise<void> {
    const { page, username, password, storageStatePathFor, dismissPostLogin } = opts;

    try {
        await page.goto('/ui/login');
        await page.getByLabel('Email Address').fill(username);
        await page.getByLabel('Password').fill(password);

        // Use `exact` as some environments also have "Login Via SSO".
        await page.getByRole('button', { name: 'LOGIN', exact: true }).click();

        // Race the success path (navigation off /ui/login) against the most common failure
        // path (rejected creds leave you on /ui/login with an inline toast). This surfaces
        // auth failures in ~15s with a clear message.
        await expect(page).not.toHaveURL(/\/ui\/login(\?|$)/, { timeout: 15_000 });
        await page.getByTestId('global_nav-dark-mode').waitFor({ state: 'visible' });

        if (dismissPostLogin) {
            await dismissPostLogin(page);
        }

        // Snapshot light first, while the UI is still in its default theme.
        await page.context().storageState({ path: storageStatePathFor('light') });

        // Click the nav dark-mode toggle. The inner Switch is `inert`, so the click handler
        // lives on the parent item targeted by this test id.
        await page.getByTestId('global_nav-dark-mode').click();

        // The store persists via a throttled (1s) subscriber. Poll until the new value lands
        // in localStorage before snapshotting storageState, so the next test boots in dark mode.
        await expect
            .poll(async () => {
                const raw = await page.evaluate(() => localStorage.getItem('persistedState'));
                return raw ? JSON.parse(raw)?.global?.view?.darkMode : null;
            })
            .toBe(true);

        await page.context().storageState({ path: storageStatePathFor('dark') });
    } catch (error) {
        throw new Error(`Auth setup failed at ${page.url()}: ${(error as Error).message}`, { cause: error });
    }
}
