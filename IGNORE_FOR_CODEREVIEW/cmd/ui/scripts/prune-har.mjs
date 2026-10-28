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

import { lstat, readdir, realpath, unlink } from 'node:fs/promises';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

function parseArguments(args) {
    let deleteOrphans = false;
    let requestedRoot;
    for (let index = 0; index < args.length; index++) {
        if (args[index] === '--delete') {
            deleteOrphans = true;
        } else if (args[index] === '--root') {
            requestedRoot = args[++index];
            if (!requestedRoot || requestedRoot.startsWith('--')) throw new Error('--root requires a directory');
        } else {
            throw new Error('Usage: yarn har:prune [--root <directory>] [--delete]');
        }
    }
    return { deleteOrphans, requestedRoot };
}

// The generated fixture belongs to the adjacent spec with the same stem.
// Directory entries that are symlinks are never followed.
export async function findOrphanedHarFiles(root) {
    let children;
    try {
        children = await readdir(root, { withFileTypes: true });
    } catch (error) {
        if (error.code === 'ENOENT') return null;
        throw error;
    }
    const candidates = await Promise.all(
        children.map(async (child) => {
            const file = path.join(root, child.name);
            if (child.isDirectory()) {
                return (await findOrphanedHarFiles(file)) ?? [];
            }
            if (!child.isFile() || !child.name.endsWith('.har')) return [];
            try {
                await lstat(file.replace(/\.har$/, '.spec.ts'));
                return [];
            } catch (error) {
                if (error.code === 'ENOENT') return [file];
                throw error;
            }
        })
    );
    return candidates.flat().sort();
}

function isInsideDirectory(directory, parent) {
    return directory.startsWith(`${parent}${path.sep}`);
}

async function pruneHarArtifacts() {
    const { deleteOrphans, requestedRoot } = parseArguments(process.argv.slice(2));
    const uiDirectory = await realpath(process.cwd());
    const root = path.resolve(uiDirectory, requestedRoot ?? 'tests/e2e');
    if (!isInsideDirectory(root, uiDirectory)) {
        throw new Error(`HAR root must be a directory inside ${uiDirectory}: ${root}`);
    }
    console.log(`Scanning ${root}.`);
    console.log(deleteOrphans ? 'Delete mode enabled.' : 'Preview mode enabled; no HAR files will be deleted.');
    let resolvedRoot;
    try {
        resolvedRoot = await realpath(root);
    } catch (error) {
        if (error.code !== 'ENOENT') throw error;
        console.log(`No HAR directory exists at ${root}.`);
        return;
    }
    if (!isInsideDirectory(resolvedRoot, uiDirectory)) {
        throw new Error(`Resolved HAR root leaves the UI workspace: ${resolvedRoot}`);
    }
    const orphans = await findOrphanedHarFiles(resolvedRoot);
    for (const file of orphans) {
        console.log(`${deleteOrphans ? 'Removing' : 'Would remove'} ${path.relative(uiDirectory, file)}`);
        if (deleteOrphans) await unlink(file);
    }
    console.log(
        `${orphans.length} orphaned HAR file${orphans.length === 1 ? '' : 's'} ${deleteOrphans ? 'removed' : 'found'}.`
    );
    if (!deleteOrphans && orphans.length)
        console.log('Run yarn har:prune --delete with the same --root to remove these files.');
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
    pruneHarArtifacts().catch((error) => {
        console.error(error);
        process.exitCode = 1;
    });
}
