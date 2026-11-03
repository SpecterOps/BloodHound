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

import { mkdir, mkdtemp, rm } from 'node:fs/promises';
import path from 'node:path';
import { runPlaywright } from '../run-specs/run-playwright.mjs';
import { makeHarCaptureArguments } from './prepare-arguments.mjs';

const uiDirectory = process.cwd();

function captureHarsByRunningSpecs(playwrightCli, spec, capturesPath) {
    runPlaywright(playwrightCli, makeHarCaptureArguments(spec.spec, uiDirectory), {
        ...process.env,
        HAR_UPDATE_CAPTURE_DIRECTORY: capturesPath,
        HAR_UPDATE_STORAGE_STATE: path.join(uiDirectory, 'playwright', '.auth', 'user-light.json'),
    });
}

export async function captureForUpdate(playwrightCli, spec) {
    const temporaryDirectory = await mkdtemp(path.join(uiDirectory, 'playwright', 'har-update-'));
    const capturesPath = path.join(temporaryDirectory, 'captures');

    await mkdir(capturesPath);

    try {
        captureHarsByRunningSpecs(playwrightCli, spec, capturesPath);

        return { captures, temporaryDirectory };
    } catch (error) {
        await rm(temporaryDirectory, { recursive: true, force: true });

        throw error;
    }
}
