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

import { existsSync } from 'node:fs';
import path from 'node:path';

const uiDirectory = process.cwd();

export function checkBrowserActionsFiles(files) {
    if (existsSync(files.spec) || existsSync(files.har)) {
        throw new Error(
            'A spec or HAR with this name already exists. Use yarn test:e2e:update-mocks to replace its HAR.'
        );
    }
}

export function announceBrowserActions(browserActions, source) {
    if (!/\bexpect\s*\(/.test(source)) {
        console.log('No assertions were recorded. Add an expected result before treating this as regression coverage.');
    }
    console.log(
        `Created ${path.relative(uiDirectory, browserActions.files.spec)} and ${path.relative(uiDirectory, browserActions.files.har)}.`
    );
    console.log('Review the assertions and captured data before committing.');
    console.log(`Run: yarn test:e2e tests/e2e/${browserActions.name}.spec.ts`);
}
