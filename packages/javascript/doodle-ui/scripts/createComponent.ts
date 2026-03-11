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

const replaceInFile = (filePath: string, searchString: string, replaceString: string) => {
    const fileContent = fs.readFileSync(filePath, 'utf8');
    const replacedContent = fileContent.replace(new RegExp(searchString, 'g'), replaceString);
    fs.writeFileSync(filePath, replacedContent, 'utf8');
};

const copyAndReplace = (sourceDir: string, destinationDir: string, searchString: string, replaceString: string) => {
    if (!fs.existsSync(destinationDir)) {
        fs.mkdirSync(destinationDir);
    } else {
        console.error(`Failed to create component. Directory ${destinationDir} already exists.`);
        process.exitCode = 1;
        return;
    }
    const files = fs.readdirSync(sourceDir);
    files.forEach((file) => {
        const sourceFilePath = path.join(sourceDir, file);
        const destFilePath = path.join(destinationDir, file.replace(new RegExp(searchString), replaceString));
        fs.copyFileSync(sourceFilePath, destFilePath);
        replaceInFile(destFilePath, searchString, replaceString);
    });
};

const generateIndex = () => {
    const componentsDir = path.join(import.meta.dirname, '../src/components');

    const exports = fs
        .readdirSync(componentsDir)
        .filter((component) => component !== 'index.ts' && component !== '__template__' && component !== 'utils.ts')
        .sort()
        .map((name) => {
            return `export * from './${name}';`;
        })
        .join('\n');

    fs.writeFileSync(path.join(componentsDir, 'index.ts'), exports + '\n');
};

const main = () => {
    const args = process.argv.slice(2);
    if (args.length < 1) {
        console.log('Usage: tsx createComponent.ts <componentName>');
        process.exitCode = 1;
        return;
    }
    const componentName = args[0];
    const sourceDir = path.join(import.meta.dirname, '../src/components/__template__');
    const destinationDir = path.join(import.meta.dirname, '../src/components', componentName);
    const searchString = '\\$0';
    const replaceString = componentName;

    copyAndReplace(sourceDir, destinationDir, searchString, replaceString);

    generateIndex();
};

main();
