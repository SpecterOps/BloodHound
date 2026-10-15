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

export function makeDevServerArguments(baseUrl) {
    const { hostname, port } = new URL(baseUrl);
    return ['dev', '--mode', 'playwright', '--host', hostname, '--port', port || '3000'];
}

async function isReachable(url) {
    try {
        const response = await fetch(url, { signal: AbortSignal.timeout(2_000) });
        await response.body?.cancel();
        return true;
    } catch {
        return false;
    }
}

async function waitForServer(url, server) {
    const deadline = Date.now() + 180_000;
    while (Date.now() < deadline) {
        if (await isReachable(url)) return;
        if (server.exitCode !== null) {
            throw new Error(`The Playwright UI server exited before ${url} became available.`);
        }
        await new Promise((resolve) => setTimeout(resolve, 500));
    }
    throw new Error(`Timed out waiting for the Playwright UI server at ${url}.`);
}

async function ensureUiServer(homeUrl) {
    if (await isReachable(homeUrl)) {
        console.log(`Reusing the UI server already running at ${homeUrl}.`);
        return undefined;
    }
    if (process.env.A11Y_TEST_SERVE?.toLowerCase() !== 'true') {
        throw new Error(
            `Cannot reach ${homeUrl}. Start the UI or set A11Y_TEST_SERVE=true in cmd/ui/.env so codegen can start it.`
        );
    }

    const serverArguments = makeDevServerArguments(process.env.A11Y_TEST_URL);
    console.log(`No UI server detected; starting the configured default: yarn ${serverArguments.join(' ')}.`);
    const yarnPath = process.env.npm_execpath;
    const server = yarnPath
        ? spawn(yarnPath, serverArguments, {
              cwd: uiDirectory,
              env: process.env,
              stdio: 'inherit',
          })
        : spawn('yarn', serverArguments, { cwd: uiDirectory, env: process.env, stdio: 'inherit' });
    try {
        await waitForServer(homeUrl, server);
    } catch (error) {
        server.kill();
        throw error;
    }
    console.log(`The UI server is ready at ${homeUrl}.`);
    return server;
}

async function ensureStorageState(playwrightCli, storageState) {
    let storageStateExists = true;
    try {
        await access(storageState);
    } catch (error) {
        if (error.code !== 'ENOENT') throw error;
        storageStateExists = false;
    }

    console.log(
        storageStateExists
            ? `Refreshing authentication state at ${path.relative(uiDirectory, storageState)} to avoid a stale session.`
            : `No authentication state found at ${path.relative(uiDirectory, storageState)}; generating it now.`
    );
    const result = spawnSync(
        process.execPath,
        [playwrightCli, 'test', '-c', 'playwright.a11y.config.ts', '--project=setup'],
        { cwd: uiDirectory, env: process.env, stdio: 'inherit' }
    );
    if (result.error) throw result.error;
    if (result.status !== 0) throw new Error(`Unable to generate authentication state (exit code ${result.status}).`);
    await access(storageState);
    console.log(`Using fresh authentication state at ${path.relative(uiDirectory, storageState)}.`);
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

    if (!requestedNames[0]) {
        console.log(`No draft name supplied; using the current branch and a random suffix: ${draftFileName}.`);
    }
    const server = await ensureUiServer(homeUrl);
    try {
        await ensureStorageState(playwrightCli, storageState);
        console.log(`Writing Playwright draft to ${path.relative(uiDirectory, output)}.`);
        console.log(`Starting codegen at the default home page: ${homeUrl}.`);
        const result = await new Promise((resolve, reject) => {
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
            child.on('error', reject);
            child.on('exit', (code, signal) => resolve({ code, signal }));
        });
        if (result.signal) process.kill(process.pid, result.signal);
        else process.exitCode = result.code ?? 1;
    } finally {
        if (server) {
            console.log('Stopping the UI server started for codegen.');
            server.kill();
        }
    }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
    run().catch((error) => {
        console.error(error.message);
        process.exitCode = 1;
    });
}
