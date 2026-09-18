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

const persistedState = {
    auth: {
        sessionToken: '',
    },
    global: {
        view: {
            darkMode: false,
            autoRunQueries: true,
            notifications: [],
            exploreLayout: 'sequential',
            isExploreTableSelected: false,
            isExploreLayoutSelected: false,
            selectedExploreTableColumns: { kind: true, label: true, objectId: true, isTierZero: true },
            pinnedExploreTableColumns: ['action-menu', 'kind', 'label'],
            timeoutSetting: false,
            isExploreGraphHighlight: true,
            isExploreGraphLabelClip: true,
        },
    },
};

export type LoginAndSnapshotThemesOptions = {
    page: Page;
    username: string;
    password: string;
    // Resolves a per-theme `storageState` file path. Callers typically wrap
    // `authStorageStateFor(theme)` from `./themes` with `path.resolve(...)`.
    storageStatePathFor: (theme: Theme) => string;
};

// Login once and snapshot `storageState` for both light and dark themes.
//
// Capturing both snapshots from a single session avoids the parallel-login race where two
// setups as the same user would invalidate each other's session. Rather than toggling the
// theme through the UI, we write the canonical light/dark `persistedState` directly into
// localStorage (carrying over the real session token from login) before each snapshot.
export async function loginAndSnapshotThemes(opts: LoginAndSnapshotThemesOptions): Promise<void> {
    const { page, username, password, storageStatePathFor } = opts;

    try {
        await page.goto('/ui/login');
        await page.getByLabel('Email Address').fill(username);
        await page.getByLabel('Password').fill(password);

        // Use `exact` as some environments also have "Login Via SSO".
        await page.getByRole('button', { name: 'LOGIN', exact: true }).click();

        // Rejected creds leave you on /ui/login, so waiting to navigate away surfaces auth
        // failures in ~15s with a clear message.
        await expect(page).not.toHaveURL(/\/ui\/login(\?|$)/, { timeout: 15_000 });

        // Carry over the real session token written by the login flow so the snapshots stay
        // authenticated after we overwrite `persistedState`.
        const sessionToken = await page.evaluate(() => {
            const raw = localStorage.getItem('persistedState');
            return raw ? JSON.parse(raw)?.auth?.sessionToken ?? '' : '';
        });

        for (const [theme, darkMode] of [
            ['light', false],
            ['dark', true],
        ] as const) {
            const state = structuredClone(persistedState);
            state.global.view.darkMode = darkMode;
            state.auth.sessionToken = sessionToken;
            await page.evaluate((next) => localStorage.setItem('persistedState', JSON.stringify(next)), state);
            await page.context().storageState({ path: storageStatePathFor(theme) });
        }
    } catch (error) {
        throw new Error(`Auth setup failed at ${page.url()}: ${(error as Error).message}`, { cause: error });
    }
}
