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

import { defineConfig, devices } from '@playwright/test';
import { authStorageStateFor, THEMES, type TestOptions } from 'bh-playwright-testing/themes';
import dotenv from 'dotenv';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
dotenv.config({ path: path.resolve(__dirname, '.env') });

// Base URL to use in actions like `await page.goto('/')`.
const baseURL = `${process.env.A11Y_TEST_URL}`;

// Address and port dev web server will run on
const { hostname, port = '3000' } = new URL(baseURL);

// When true, browser will run tests on Vite web server (API service must still be running)
// Set to false when testing against http://bloodhound.localhost or https://test.bloodhoundenterprise.io
const shouldRunWebServer = process.env.A11Y_TEST_SERVE?.toLowerCase() === 'true';

const a11yTestMatch = /.*\.a11y\.spec\.ts/;

// Browser dimension of the project matrix. Pair each with each theme below to produce e.g.
// `chromium-light`, `chromium-dark`, `firefox-light`, `firefox-dark`.
const browsers = [
    { name: 'chromium', device: devices['Desktop Chrome'] },
    { name: 'firefox', device: devices['Desktop Firefox'] },
] as const;

const webServer = {
    command: `yarn dev --host ${hostname} --port ${port}`,
    url: baseURL,
    reuseExistingServer: !process.env.CI,
    timeout: 180_000,
};

export default defineConfig<TestOptions>({
    testDir: './tests/a11y',
    outputDir: './playwright/a11y/results',
    fullyParallel: true,
    // Fails the build on CI if "test.only" left in source code.
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 1 : 0,
    workers: process.env.CI ? 1 : undefined,
    timeout: 60_000,
    expect: {
        timeout: 10_000,
    },
    reporter: [
        [process.env.CI ? 'line' : 'list'],
        [
            'html',
            {
                open: 'never',
                outputFolder: './playwright/a11y/html-report',
            },
        ],
        [
            'allure-playwright',
            {
                resultsDir: './playwright/a11y/allure-results',
            },
        ],
    ],
    use: {
        ...devices['Desktop Chrome'],
        baseURL,
        screenshot: 'only-on-failure',
        trace: 'retain-on-failure',
        // Block service workers so MSW (registered by the Vite dev build in `main.tsx`) cannot
        // take over network traffic. With MSW active, requests are routed through its Service
        // Worker before Playwright sees them, which makes `page.route` / `context.route` invisible
        // to API calls and causes test-only stubs (e.g. the `useGraphHasData` cypher stub
        // installed in `global.setup.ts` and `tests/a11y/fixtures.ts`) to silently fall through to
        // the real backend. `main.tsx` swallows the resulting registration failure so the app
        // still mounts.
        serviceWorkers: 'block',
    },
    // Browser × theme matrix. A single `setup` project logs in once and snapshots both
    // light and dark storage states (see global.setup.ts); each browser-theme project then loads
    // its matching snapshot. Running setup once avoids a race where two parallel logins as the
    // same user would invalidate each other's session. Project names follow `<browser>-<theme>`
    // so report grouping and CLI filtering (e.g. `--project=chromium-dark`) work naturally.
    projects: [
        {
            name: 'setup',
            testDir: './tests',
            testMatch: /global\.setup\.ts$/,
        },
        ...THEMES.flatMap((theme) =>
            browsers.map(({ name, device }) => ({
                name: `${name}-${theme}`,
                testMatch: a11yTestMatch,
                use: {
                    ...device,
                    storageState: authStorageStateFor(theme),
                    theme,
                },
                dependencies: ['setup'],
            }))
        ),
    ],
    ...(shouldRunWebServer ? { webServer } : {}),
});
