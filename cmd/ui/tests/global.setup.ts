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

import { test as setup } from 'bh-playwright-testing';
import { loginAndSnapshotThemes } from 'bh-playwright-testing/auth';
import { installGraphHasDataStub } from 'bh-playwright-testing/stubs/graph-has-data';
import { authStorageStateFor, type Theme } from 'bh-playwright-testing/themes';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const username = process.env.A11Y_TEST_USERNAME;
const password = process.env.A11Y_TEST_PASSWORD;

if (!username || !password) {
    throw new Error('A11Y_TEST_USERNAME and A11Y_TEST_PASSWORD must be set (see .env.example).');
}

setup('Generate and cache auth state for light and dark theme', async ({ page }) => {
    // Install the cypher stub before navigation so `useGraphHasData` resolves to "true" and
    // the "No Data Available" upload dialog stays closed during login.
    await installGraphHasDataStub(page);

    await loginAndSnapshotThemes({
        page,
        username,
        password,
        storageStatePathFor: (theme: Theme) => path.resolve(__dirname, '..', authStorageStateFor(theme)),
    });
});
