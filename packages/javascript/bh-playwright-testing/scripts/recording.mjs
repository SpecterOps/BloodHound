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
import { existsSync } from 'node:fs';
import { access, mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';
import { parseArgs } from 'node:util';
import prettier from 'prettier';
import { makeMockedTest, sanitizeRecording } from './mocked-codegen.mjs';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

export function parseRecordingArguments(args) {
    const usage = 'Usage: yarn codegen <name> [--url /ui/feature]';
    let parsed;
    try {
        parsed = parseArgs({
            args,
            options: { url: { type: 'string' } },
            allowPositionals: true,
        });
    } catch {
        throw new Error(usage);
    }
    const [name] = parsed.positionals;
    const requestedUrl = parsed.values.url ?? '/ui/';
    if (parsed.positionals.length !== 1 || !/^[a-z0-9][a-z0-9_-]*$/i.test(name) || !requestedUrl) {
        throw new Error(usage);
    }
    return { name, requestedUrl };
}

export function checkRecordingFiles(files) {
    if (existsSync(files.spec) || existsSync(files.har)) {
        throw new Error(
            'A spec or HAR with this name already exists. Use yarn test:e2e:update-mocks to replace its HAR.'
        );
    }
}

export async function saveRecording(files, har, spec) {
    checkRecordingFiles(files);
    await mkdir(path.dirname(files.har), { recursive: true });
    await writeFile(files.har, `${JSON.stringify(har, null, 2)}\n`, { flag: 'wx' });
    await writeFile(files.spec, spec, { flag: 'wx' });
}

export function makeDevServerArguments(baseUrl) {
    const { hostname, port } = new URL(baseUrl);
    return ['dev', '--mode', 'playwright', '--host', hostname, '--port', port || '3000'];
}

export function recorderArguments(storageState, source, har, url) {
    return [
        'codegen',
        '--block-service-workers',
        '--load-storage',
        storageState,
        '--output',
        source,
        '--save-har',
        har,
        '--save-har-glob',
        '**/api/**',
        '--viewport-size',
        '1280,720',
        url,
    ];
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
    console.log(`Refreshing authentication state at ${path.relative(uiDirectory, storageState)}.`);
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

export async function runRecording() {
    const { name, requestedUrl } = parseRecordingArguments(process.argv.slice(2));
    if (!process.env.A11Y_TEST_URL) throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');
    const outputDirectory = path.join(uiDirectory, 'tests', 'e2e');
    const files = {
        spec: path.join(outputDirectory, `${name}.spec.ts`),
        har: path.join(outputDirectory, `${name}.har`),
    };
    checkRecordingFiles(files);
    const storageState = path.join(uiDirectory, 'playwright', '.auth', 'user-light.json');
    const homeUrl = new URL(requestedUrl, process.env.A11Y_TEST_URL).toString();
    if (new URL(homeUrl).origin !== new URL(process.env.A11Y_TEST_URL).origin) {
        throw new Error('--url must use the configured A11Y_TEST_URL origin.');
    }
    const require = createRequire(path.join(uiDirectory, 'package.json'));
    const playwrightCli = require.resolve('@playwright/test/cli');

    await mkdir(path.join(uiDirectory, 'playwright'), { recursive: true });
    const temporaryDirectory = await mkdtemp(path.join(uiDirectory, 'playwright', 'codegen-'));
    const rawSource = path.join(temporaryDirectory, 'recording.ts');
    const rawHar = path.join(temporaryDirectory, 'recording.har');
    const server = await ensureUiServer(homeUrl);
    try {
        await ensureStorageState(playwrightCli, storageState);
        console.log(`Opening the Playwright browser at ${homeUrl}; this may take a few seconds...`);
        console.log('Record the flow, add assertions, then close the browser to save.');
        const result = await new Promise((resolve, reject) => {
            const child = spawn(
                process.execPath,
                [playwrightCli, ...recorderArguments(storageState, rawSource, rawHar, homeUrl)],
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
            throw new Error(
                `Recording did not finish successfully. Temporary recording files remain in ${temporaryDirectory}.`
            );
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
            ...(await prettier.resolveConfig(files.spec)),
            parser: 'typescript',
        });
        await saveRecording(files, har, spec);
        await rm(temporaryDirectory, { recursive: true });
        if (!/\bexpect\s*\(/.test(source)) {
            console.log(
                'No assertions were recorded. Add an expected result before treating this as regression coverage.'
            );
        }
        console.log(`Created ${path.relative(uiDirectory, files.spec)} and ${path.relative(uiDirectory, files.har)}.`);
        console.log('Review the assertions and captured data before committing.');
        console.log(`Run: yarn test:e2e tests/e2e/${name}.spec.ts`);
        console.log('Checking the new test with API replay and no live login...');
        const replay = spawnSync(
            process.execPath,
            [
                playwrightCli,
                'test',
                '-c',
                fileURLToPath(new URL('../playwright.e2e.config.mjs', import.meta.url)),
                `tests/e2e/${name}.spec.ts`,
                '--project=chromium-light',
            ],
            { cwd: uiDirectory, env: process.env, stdio: 'inherit' }
        );
        if (replay.error) throw replay.error;
        process.exitCode = replay.status ?? 1;
        console.log(
            replay.status === 0
                ? 'Mocked test passed. Review the spec and HAR before committing.'
                : 'Replay failed; the files are saved. Review the failure and record again if needed.'
        );
    } finally {
        if (server) {
            console.log('Stopping the UI server started for codegen.');
            server.kill();
        }
    }
}
