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
import { spawn } from 'node:child_process';
import { createRequire } from 'node:module';
import path from 'node:path';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

const args = process.argv.slice(2);
const isHelp = args.includes('--help') || args.includes('-h');
const hasUrl = args.some((arg, index) => /^https?:\/\//i.test(arg) && args[index - 1] !== '--proxy-server');
if (!isHelp && !hasUrl) {
    if (!process.env.A11Y_TEST_URL) {
        console.error('Set A11Y_TEST_URL in cmd/ui/.env or pass a URL to yarn codegen.');
        process.exitCode = 1;
    } else {
        args.push(new URL('/ui/login', process.env.A11Y_TEST_URL).toString());
    }
}

if (!process.exitCode) {
    const require = createRequire(path.join(uiDirectory, 'package.json'));
    const playwrightCli = require.resolve('@playwright/test/cli');
    const child = spawn(process.execPath, [playwrightCli, 'codegen', '--block-service-workers', ...args], {
        cwd: uiDirectory,
        env: process.env,
        stdio: 'inherit',
    });
    child.on('error', (error) => {
        console.error(`Unable to start Playwright codegen: ${error.message}`);
        process.exitCode = 1;
    });
    child.on('exit', (code, signal) => {
        if (signal) process.kill(process.pid, signal);
        else process.exitCode = code ?? 1;
    });
}
