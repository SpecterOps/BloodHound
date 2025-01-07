// Copyright 2024 Specter Ops, Inc.
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

import * as fs from 'node:fs';

function createSupportFolders(path: string): void {
    if (!fs.existsSync(path)) {
        fs.mkdirSync(path, { recursive: true });
        console.log(`${path} folder is created`);
    } else {
        console.log(`${path} folder exists`);
    }
}

function createRerunFile(fileName: string): void {
    if (!fs.existsSync(fileName)) {
        fs.writeFileSync(fileName, '');
        console.log(`${fileName} is created`);
    } else {
        console.log(`${fileName} file exists`);
    }
}

const reRunFile = '@rerun.txt';
createRerunFile(reRunFile);

const reportPath = './test-results/reports';
createSupportFolders(reportPath);

const screenshotsPath = './test-results/screenshots';
createSupportFolders(screenshotsPath);
