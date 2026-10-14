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
import { spawn, spawnSync } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import { access } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

export function makeDraftFileName(requestedName, branchName, randomSuffix) {
    if (requestedName && path.basename(requestedName) !== requestedName) {
        throw new Error('Draft name must be a file name without a directory.');
    }

    const requestedStem = requestedName?.replace(/\.ts$/, '');
    const generatedStem = `${branchName || 'playwright-draft'}-${randomSuffix}`;
    const stem = (requestedStem || generatedStem)
        .normalize('NFKD')
        .replace(/[^a-zA-Z0-9_-]+/g, '-')
        .replace(/^-+|-+$/g, '');
    if (!stem) throw new Error('Draft name must contain at least one letter or number.');
    return `${stem}.ts`;
}

function currentBranch() {
    const result = spawnSync('git', ['branch', '--show-current'], { cwd: uiDirectory, encoding: 'utf8' });
    return result.status === 0 ? result.stdout.trim() : '';
}

async function ensureStorageState(playwrightCli, storageState) {
    try {
        await access(storageState);
        return;
    } catch (error) {
        if (error.code !== 'ENOENT') throw error;
    }

    console.log(`No authentication state found at ${path.relative(uiDirectory, storageState)}; generating it now.`);
    const result = spawnSync(
        process.execPath,
        [playwrightCli, 'test', '-c', 'playwright.a11y.config.ts', '--project=setup'],
        { cwd: uiDirectory, env: process.env, stdio: 'inherit' }
    );
    if (result.error) throw result.error;
    if (result.status !== 0) throw new Error(`Unable to generate authentication state (exit code ${result.status}).`);
    await access(storageState);
}

async function run() {
    const requestedNames = process.argv.slice(2);
    if (requestedNames.length > 1) throw new Error('Usage: yarn codegen [draft-name]');
    if (!process.env.A11Y_TEST_URL) throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');

    const draftFileName = makeDraftFileName(requestedNames[0], currentBranch(), randomBytes(4).toString('hex'));
    const output = path.join(uiDirectory, 'playwright', draftFileName);
    const storageState = path.join(uiDirectory, 'playwright', '.auth', 'user-light.json');
    const homeUrl = new URL('/ui/', process.env.A11Y_TEST_URL).toString();
    const require = createRequire(path.join(uiDirectory, 'package.json'));
    const playwrightCli = require.resolve('@playwright/test/cli');

    await ensureStorageState(playwrightCli, storageState);
    console.log(`Writing Playwright draft to ${path.relative(uiDirectory, output)}.`);
    const child = spawn(
        process.execPath,
        [
            playwrightCli,
            'codegen',
            '--block-service-workers',
            '--load-storage',
            storageState,
            '--output',
            output,
            homeUrl,
        ],
        {
            cwd: uiDirectory,
            env: process.env,
            stdio: 'inherit',
        }
    );
    child.on('error', (error) => {
        console.error(`Unable to start Playwright codegen: ${error.message}`);
        process.exitCode = 1;
    });
    child.on('exit', (code, signal) => {
        if (signal) process.kill(process.pid, signal);
        else process.exitCode = code ?? 1;
    });
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
    run().catch((error) => {
        console.error(error.message);
        process.exitCode = 1;
    });
}
