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

import dotenv from 'dotenv';
import { mkdir, mkdtemp, readFile, rm } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { selectFailingTests } from '../update-hars/select-failing-tests.mjs';
import { listRecordedTests } from '../update-hars/select-tests.mjs';
import { updateSelectedTests } from '../update-hars/index.mjs';
import { runSuite } from './suite.mjs';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

async function readFailures(reportFile, tests) {
    let report;
    try {
        report = JSON.parse(await readFile(reportFile, 'utf8'));
    } catch {
        throw new Error('E2E tests failed before Playwright could report the failing specs.');
    }
    const failed = selectFailingTests(report, tests);
    if (!failed.length) throw new Error('E2E tests failed, but no recorded spec was identified as failing.');
    return failed;
}

async function main() {
    if (process.argv.length !== 2) throw new Error('Usage: yarn test:e2e:update-failing-mocks');
    const tests = await listRecordedTests(path.join(uiDirectory, 'tests', 'e2e'));
    const playwrightCli = createRequire(path.join(uiDirectory, 'package.json')).resolve('@playwright/test/cli');
    await mkdir(path.join(uiDirectory, 'playwright'), { recursive: true });
    const temporaryDirectory = await mkdtemp(path.join(uiDirectory, 'playwright', 'har-failures-'));
    try {
        const reportFile = path.join(temporaryDirectory, 'results.json');
        if (runSuite(playwrightCli, tests, reportFile) === 0) {
            console.log('All recorded tests passed; no HARs need updating.');
            return;
        }
        const failed = await readFailures(reportFile, tests);
        console.log(`Updating HARs for: ${failed.map((test) => test.name).join(', ')}.`);
        await updateSelectedTests(failed);
        if (runSuite(playwrightCli, failed) !== 0) throw new Error('Updated HARs did not pass replay.');
    } finally {
        await rm(temporaryDirectory, { recursive: true, force: true });
    }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
    main().catch((error) => {
        console.error(error.message);
        process.exitCode = 1;
    });
}
