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
import dotenv from 'dotenv';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

dotenv.config({ path: path.dirname(fileURLToPath(import.meta.url)) + '/.env', quiet: true });
const baseURL = process.env.A11Y_TEST_URL || 'http://127.0.0.1:3000';
const { hostname, port } = new URL(baseURL);

export default defineConfig({
    testDir: './tests/e2e',
    outputDir: './playwright/e2e/results',
    reporter: process.env.CI ? 'line' : 'list',
    use: { baseURL, serviceWorkers: 'block' },
    projects: [{ name: 'chromium-light', use: devices['Desktop Chrome'] }],
    webServer:
        process.env.A11Y_TEST_SERVE?.toLowerCase() === 'true'
            ? {
                  command: `yarn dev --mode playwright --host ${hostname} --port ${port || '3000'}`,
                  url: baseURL,
                  reuseExistingServer: !process.env.CI,
                  timeout: 180_000,
              }
            : undefined,
});
