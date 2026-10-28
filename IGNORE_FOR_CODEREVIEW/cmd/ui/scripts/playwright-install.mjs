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
import { createRequire } from 'node:module';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const uiDirectory = process.cwd();

export function resolveBrowsers(requestedBrowsers) {
    return requestedBrowsers.length ? requestedBrowsers : ['chromium', 'firefox'];
}

function run() {
    const requestedBrowsers = process.argv.slice(2);
    const browsers = resolveBrowsers(requestedBrowsers);
    if (!requestedBrowsers.length) {
        console.log(`No browsers supplied; installing the default browsers: ${browsers.join(', ')}.`);
    }

    const require = createRequire(path.join(uiDirectory, 'package.json'));
    const playwrightCli = require.resolve('@playwright/test/cli');
    const child = spawn(process.execPath, [playwrightCli, 'install', ...browsers], {
        cwd: uiDirectory,
        env: process.env,
        stdio: 'inherit',
    });

    child.on('error', (error) => {
        console.error(`Unable to install Playwright browsers: ${error.message}`);
        process.exitCode = 1;
    });
    child.on('exit', (code, signal) => {
        if (signal) process.kill(process.pid, signal);
        else process.exitCode = code ?? 1;
    });
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) run();
