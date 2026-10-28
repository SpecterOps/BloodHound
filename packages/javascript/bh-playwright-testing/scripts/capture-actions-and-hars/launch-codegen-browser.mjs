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

import { spawn } from 'node:child_process';
import { makeDevServerArguments, recorderArguments } from './prepare-arugments.mjs';

const uiDirectory = process.cwd();

export async function launchCodegenBrowser(browserActions) {
    console.log(`Opening the Playwright browser at ${browserActions.homeUrl}; this may take a few seconds...`);
    console.log('Record the flow, add assertions, then close the browser to save.');
    const result = await new Promise((resolve, reject) => {
        const child = spawn(
            process.execPath,
            [
                browserActions.playwrightCli,
                ...recorderArguments(
                    browserActions.storageState,
                    browserActions.rawSource,
                    browserActions.rawHar,
                    browserActions.homeUrl
                ),
            ],
            { cwd: uiDirectory, env: process.env, stdio: 'inherit' }
        );
        child.on('error', reject);
        child.on('exit', (code, signal) => resolve({ code, signal }));
    });
    if (result.signal || result.code !== 0)
        throw new Error(
            `Browser action capture did not finish successfully. Temporary files remain in ${browserActions.temporaryDirectory}.`
        );
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
        if (server.exitCode !== null)
            throw new Error(`The Playwright UI server exited before ${url} became available.`);
        await new Promise((resolve) => setTimeout(resolve, 500));
    }
    throw new Error(`Timed out waiting for the Playwright UI server at ${url}.`);
}

function startUiServer(baseUrl) {
    const serverArguments = makeDevServerArguments(baseUrl);
    console.log(`No UI server detected; starting the configured default: yarn ${serverArguments.join(' ')}.`);
    const yarnPath = process.env.npm_execpath;
    return yarnPath
        ? spawn(yarnPath, serverArguments, { cwd: uiDirectory, env: process.env, stdio: 'inherit' })
        : spawn('yarn', serverArguments, { cwd: uiDirectory, env: process.env, stdio: 'inherit' });
}

export async function prepareUiServer(homeUrl) {
    if (await isReachable(homeUrl)) {
        console.log(`Reusing the UI server already running at ${homeUrl}.`);
        return undefined;
    }
    if (process.env.A11Y_TEST_SERVE?.toLowerCase() !== 'true') {
        throw new Error(
            `Cannot reach ${homeUrl}. Start the UI or set A11Y_TEST_SERVE=true in cmd/ui/.env so codegen can start it.`
        );
    }
    const server = startUiServer(process.env.A11Y_TEST_URL);
    try {
        await waitForServer(homeUrl, server);
    } catch (error) {
        server.kill();
        throw error;
    }
    console.log(`The UI server is ready at ${homeUrl}.`);
    return server;
}
