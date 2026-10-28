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

import { defineConfig } from '@playwright/test';
import type { A11yTestOptions } from 'bh-playwright-testing';
import browserConfig from './playwright.a11y.config';

export default defineConfig<A11yTestOptions>({
    ...browserConfig,
    testDir: './tests/e2e',
    testMatch: '**/*.spec.ts',
    outputDir: './playwright/e2e/results',
    reporter: [
        [process.env.CI ? 'line' : 'list'],
        ['html', { open: 'never', outputFolder: './playwright/e2e/html-report' }],
    ],
    projects: browserConfig.projects
        ?.filter((project) => project.name !== 'setup')
        .map((project) => ({
            ...project,
            testMatch: '**/*.spec.ts',
            dependencies: [],
            use: { ...project.use, storageState: { cookies: [], origins: [] } },
        })),
});
