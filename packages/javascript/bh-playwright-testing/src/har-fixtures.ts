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

import { test as base, expect } from '@playwright/test';
import { configureHarCapture, configureHarReplay } from './configure-har-fixtures';

/** Self-contained API replay for generated tests, with an isolated fake session. */
export const test = base.extend<{ replayHar: URL | undefined }>({
    replayHar: [undefined, { option: true }],
    serviceWorkers: 'block',
    storageState: async ({ browserName }, use) => {
        void browserName;

        if (process.env.HAR_UPDATE_CAPTURE_DIRECTORY) {
            if (!process.env.HAR_UPDATE_STORAGE_STATE) {
                throw new Error('HAR update requires authentication state.');
            }

            await use(process.env.HAR_UPDATE_STORAGE_STATE);
        } else {
            await use({ cookies: [], origins: [] });
        }
    },
    context: async ({ context, baseURL, replayHar }, use, testInfo) => {
        if (!replayHar) {
            throw new Error('Set replayHar to the recorded HAR URL with test.use().');
        }

        if (!baseURL) {
            throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');
        }

        if (process.env.HAR_UPDATE_CAPTURE_DIRECTORY) {
            await configureHarCapture(context, process.env.HAR_UPDATE_CAPTURE_DIRECTORY, testInfo);
            await use(context);

            return;
        }

        await configureHarReplay(context, baseURL, replayHar, testInfo);

        await use(context);
    },
});

export { expect };
