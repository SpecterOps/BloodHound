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
import { spawnSync } from 'node:child_process';
import { access, lstat, mkdtemp, readFile, readdir, realpath, rm, rmdir, unlink } from 'node:fs/promises';
import { createRequire } from 'node:module';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

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

async function readSuiteReport() {
    const listingDirectory = await mkdtemp(path.join(tmpdir(), 'bh-har-list-'));
    const listingFile = path.join(listingDirectory, 'tests.json');
    let report;
    try {
        const require = createRequire(path.join(uiDirectory, 'package.json'));
        const cli = require.resolve('@playwright/test/cli');
        const result = spawnSync(
            process.execPath,
            [cli, 'test', '-c', 'playwright.a11y.config.ts', '--list', '--reporter=json'],
            {
                cwd: uiDirectory,
                env: { ...process.env, PLAYWRIGHT_JSON_OUTPUT_FILE: listingFile },
                encoding: 'utf8',
            }
        );
        if (result.error || result.status !== 0) {
            throw new Error(
                `Unable to list the complete Playwright suite: ${result.stderr || result.stdout || result.error}`
            );
        }
        report = JSON.parse(await readFile(listingFile, 'utf8'));
    } finally {
        await unlink(listingFile).catch(() => {});
        await rmdir(listingDirectory);
    }

    if (!report.config?.rootDir || report.errors?.length || !report.suites?.length) {
        throw new Error('Playwright did not return a complete suite; no HAR files were touched.');
    }
    return report;
}

function collectActiveTestIds(suites, activeIds = new Set()) {
    for (const suite of suites) {
        for (const spec of suite.specs ?? []) activeIds.add(spec.id);
        collectActiveTestIds(suite.suites ?? [], activeIds);
    }
    return activeIds;
}

async function pruneHarArtifacts() {
    const { deleteOrphans, requestedRoot } = parseArguments(process.argv.slice(2));
    const report = await readSuiteReport();
    const activeIds = collectActiveTestIds(report.suites);
    const configRoot = path.resolve(report.config.rootDir);
    const root = path.resolve(configRoot, requestedRoot ?? 'test-artifacts/har');
    if (!isInsideDirectory(root, uiDirectory)) {
        throw new Error(`HAR root must be a directory inside ${uiDirectory}: ${root}`);
    }
    const orphans = await findOrphanedHarDirectories(root, configRoot, activeIds);
    if (orphans === null) {
        console.log(`No HAR artifact directory exists at ${root}.`);
        return;
    }
    const resolvedRoot = await realpath(root);
    if (!isInsideDirectory(resolvedRoot, uiDirectory)) {
        throw new Error(`Resolved HAR root leaves the UI workspace: ${resolvedRoot}`);
    }

    for (const directory of orphans) {
        console.log(`${deleteOrphans ? 'Removing' : 'Would remove'} ${path.relative(uiDirectory, directory)}`);
        if (deleteOrphans) await rm(directory, { recursive: true });
    }
    console.log(
        `${orphans.length} orphaned HAR director${orphans.length === 1 ? 'y' : 'ies'} ${deleteOrphans ? 'removed' : 'found'}.`
    );
    if (!deleteOrphans && orphans.length) console.log('Run yarn har:prune --delete to remove these directories.');
}

// Paths passed here have already been resolved. Requiring the separator excludes both
// the parent itself and sibling names sharing its prefix.
function isInsideDirectory(directory, parent) {
    return directory.startsWith(`${parent}${path.sep}`);
}

export async function findOrphanedHarDirectories(root, configRoot, activeIds) {
    let children;
    try {
        children = await readdir(root, { withFileTypes: true });
    } catch (error) {
        if (error.code === 'ENOENT') return null;
        throw error;
    }

    const resolvedRoot = await realpath(root);
    const orphans = [];
    for (const child of children) {
        if (!child.isDirectory()) continue;
        const directory = path.join(root, child.name);
        const markerFile = path.join(directory, '.har-fixture.json');
        let marker;
        try {
            marker = JSON.parse(await readFile(markerFile, 'utf8'));
        } catch {
            continue; // Unmarked or malformed directories are never deleted automatically.
        }
        const belongsToSuite = marker.rootDir === configRoot;
        const hasTestIdentity = typeof marker.testId === 'string';
        if (!belongsToSuite || !hasTestIdentity || activeIds.has(marker.testId)) {
            continue;
        }
        try {
            await access(path.join(directory, 'recording.har'));
        } catch {
            continue;
        }
        if (!(await lstat(directory)).isDirectory() || !isInsideDirectory(await realpath(directory), resolvedRoot)) {
            continue;
        }
        orphans.push(directory);
    }
    return orphans;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
    pruneHarArtifacts().catch((error) => {
        console.error(error);
        process.exitCode = 1;
    });
}
