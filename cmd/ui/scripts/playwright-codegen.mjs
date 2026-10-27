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
import { access, mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';
import prettier from 'prettier';
import { makeMockedTest, sanitizeRecording } from './playwright-mocked-codegen.mjs';

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
    const args = process.argv.slice(2);
    const urlIndex = args.indexOf('--url');
    const requestedUrl = urlIndex < 0 ? '/ui/' : args[urlIndex + 1];
    if (urlIndex >= 0) args.splice(urlIndex, 2);
    if (!requestedUrl || args.length > 1 || args.some((arg) => arg.startsWith('--'))) {
        throw new Error('Usage: yarn codegen [test-name] [--url /ui/feature]');
    }
    if (!process.env.A11Y_TEST_URL) throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');

    const draftFileName = makeDraftFileName(args[0], currentBranch(), randomBytes(4).toString('hex'));
    const name = draftFileName.replace(/(?:\.a11y\.spec)?\.ts$/, '');
    const outputDirectory = path.join(uiDirectory, 'tests', 'e2e');
    const output = path.join(outputDirectory, `${name}.spec.ts`);
    const fixture = path.join(outputDirectory, `${name}.har`);
    const storageState = path.join(uiDirectory, 'playwright', '.auth', 'user-light.json');
    const homeUrl = new URL(requestedUrl, process.env.A11Y_TEST_URL).toString();
    if (new URL(homeUrl).origin !== new URL(process.env.A11Y_TEST_URL).origin) {
        throw new Error('--url must use the configured A11Y_TEST_URL origin.');
    }
    const require = createRequire(path.join(uiDirectory, 'package.json'));
    const playwrightCli = require.resolve('@playwright/test/cli');

    for (const target of [output, fixture]) {
        try {
            await access(target);
        } catch (error) {
            if (error.code === 'ENOENT') continue;
            throw error;
        }
        throw new Error(`File already exists: ${target}. Choose a new test name.`);
    }
    await mkdir(path.join(uiDirectory, 'playwright'), { recursive: true });
    const temporaryDirectory = await mkdtemp(path.join(uiDirectory, 'playwright', 'codegen-'));
    const rawSource = path.join(temporaryDirectory, 'recording.ts');
    const rawHar = path.join(temporaryDirectory, 'recording.har');
    if (!args[0]) {
        console.log(`No draft name supplied; using the current branch and a random suffix: ${draftFileName}.`);
    }
    const server = await ensureUiServer(homeUrl);
    try {
        await ensureStorageState(playwrightCli, storageState);
        console.log(`Recording ${homeUrl}. Click through your feature and use the recorder's assertion tools.`);
        console.log('Close the recording browser when finished to save the test and API mocks.');
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
                    rawSource,
                    '--save-har',
                    rawHar,
                    '--save-har-glob',
                    '**/api/**',
                    '--viewport-size',
                    '1280,720',
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
        if (result.signal || result.code !== 0) {
            throw new Error(`Recording did not finish successfully. Draft files remain in ${temporaryDirectory}.`);
        }
        const source = await readFile(rawSource, 'utf8');
        const har = sanitizeRecording(
            JSON.parse(await readFile(rawHar, 'utf8')),
            JSON.parse(await readFile(storageState, 'utf8'))
        );
        if (!har.log.entries.length)
            throw new Error('No API responses captured. Visit your feature before closing codegen.');
        const licenseHeader = (await readFile(fileURLToPath(import.meta.url), 'utf8')).split('\n\n')[0];
        const spec = await prettier.format(makeMockedTest(source, name, process.env.A11Y_TEST_URL, licenseHeader), {
            ...(await prettier.resolveConfig(output)),
            parser: 'typescript',
        });
        await mkdir(outputDirectory, { recursive: true });
        await writeFile(fixture, `${JSON.stringify(har, null, 2)}\n`, { flag: 'wx' });
        await writeFile(output, spec, { flag: 'wx' });
        await rm(temporaryDirectory, { recursive: true });
        if (!/\bexpect\s*\(/.test(source)) {
            console.log(
                'No assertions were recorded. Add an expected result before treating this as regression coverage.'
            );
        }
        console.log(`Created ${path.relative(uiDirectory, output)} and ${path.relative(uiDirectory, fixture)}.`);
        console.log('Review the assertions and captured data before committing.');
        console.log(`Run: yarn test:e2e tests/e2e/${name}.spec.ts --project=chromium-light`);
        console.log('Checking the new test with API replay and no live login...');
        const replay = spawnSync(
            process.execPath,
            [
                playwrightCli,
                'test',
                '-c',
                'playwright.e2e.config.ts',
                `tests/e2e/${name}.spec.ts`,
                '--project=chromium-light',
            ],
            { cwd: uiDirectory, env: process.env, stdio: 'inherit' }
        );
        if (replay.error) throw replay.error;
        process.exitCode = replay.status ?? 1;
        console.log(
            replay.status === 0
                ? 'Mocked test passed. Review and commit the generated spec and HAR together.'
                : 'Replay failed; the generated files are saved. Review the failure and adjust the test or re-record.'
        );
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
