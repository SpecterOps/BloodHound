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
import fs from 'fs';
import path from 'path';

const urlVersion = '__URL_VERSION__';
const version = '__VERSION__';
const templatePrefix = '<img src="https://img.shields.io/badge/version-';

const badgeTemplate = `${templatePrefix}${urlVersion}-teal" alt="version ${version}"/>`;

const main = () => {
    const args = process.argv.slice(2);
    const versionString = args[0];
    const readmePath = path.join(args[1], 'README.md');

    let newBadgeTag = badgeTemplate.replace(version, versionString);
    newBadgeTag = newBadgeTag.replace(urlVersion, versionString.replace('-', '--'));

    const readmeLines = fs.readFileSync(readmePath, 'utf-8').split('\n');

    let versionBadgeLineIndex = -1;
    readmeLines.forEach((line, index) => {
        if (line.includes(templatePrefix)) versionBadgeLineIndex = index;
        return;
    });

    if (versionBadgeLineIndex < 0) {
        console.warn(`Did not find the img tag for replacing the version in.`);
        return;
    }

    if (readmeLines[versionBadgeLineIndex] !== newBadgeTag) {
        console.log(
            `
            
Replacing the version badge in the README with version ${versionString}. Please make sure to commit the changes!
            
            `
        );
        readmeLines[versionBadgeLineIndex] = newBadgeTag;
        fs.writeFileSync(readmePath, readmeLines.join('\n'));
    }
};

main();
