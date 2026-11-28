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

import { mkdir, mkdtemp } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import { checkForDuplicateSpecName } from './manage-files.mjs';
import { parseRecordingSessionArguments } from './prepare-arguments.mjs';

export function makeRecordingFilePaths(uiDirectory, name) {
    return {
        spec: path.join(uiDirectory, 'tests', 'e2e', `${name}.spec.ts`),
        har: path.join(uiDirectory, 'tests', 'e2e', `${name}.har`),
    };
}

export function checkHasSameOrigin(homeUrl, baseUrl) {
    return new URL(homeUrl).origin === new URL(baseUrl).origin;
}

export function makeRecordingSession({ name, files, homeUrl, playwrightCli, temporaryDirectory, uiDirectory }) {
    return {
        name,
        files,
        homeUrl,
        playwrightCli,
        temporaryDirectory,
        storageState: path.join(uiDirectory, 'playwright', '.auth', 'user-light.json'),
        rawSource: path.join(temporaryDirectory, 'recordingSession.ts'),
        rawHar: path.join(temporaryDirectory, 'capture.har'),
    };
}

export async function prepareCodegenRecordingSession(args) {
    const uiDirectory = process.cwd();
    const baseUrl = process.env.A11Y_TEST_URL;
    const { name, requestedUrl } = parseRecordingSessionArguments(args);

    if (!baseUrl) {
        throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');
    }

    const recordingFiles = makeRecordingFilePaths(uiDirectory, name);
    checkForDuplicateSpecName(recordingFiles);

    const homeUrl = new URL(requestedUrl, baseUrl).toString();
    const sameOrigin = checkHasSameOrigin(homeUrl, baseUrl);

    if (!sameOrigin) {
        throw new Error('--url must use the configured A11Y_TEST_URL origin.');
    }

    const packageJsonPath = path.join(uiDirectory, 'package.json');
    const requireFromUi = createRequire(packageJsonPath);
    const playwrightCli = requireFromUi.resolve('@playwright/test/cli');
    const playwrightDirectory = path.join(uiDirectory, 'playwright');
    const temporaryDirectoryPrefix = path.join(playwrightDirectory, 'codegen-');

    await mkdir(playwrightDirectory, { recursive: true });
    const temporaryDirectory = await mkdtemp(temporaryDirectoryPrefix);
    const recordingSession = makeRecordingSession({
        name,
        files: recordingFiles,
        homeUrl,
        playwrightCli,
        temporaryDirectory,
        uiDirectory,
    });

    return recordingSession;
}
