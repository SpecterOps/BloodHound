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

import { createHash } from 'node:crypto';
import { mkdir, readFile, rename, writeFile } from 'node:fs/promises';
import path from 'node:path';
import type { TestInfo } from '@playwright/test';

export type HarMode = 'record' | 'update' | 'mock' | 'off';
export type HarNotFound = 'abort' | 'fallback';

export type HarArtifactPaths = {
    directory: string;
    recording: string;
    requests: string;
    responses: string;
};

export type HarRequestSummary = { method: string; url: string; bodySize: number | null };
export type HarResponseSummary = {
    method: string;
    url: string;
    status: number;
    mimeType: string;
    bodySize: number | null;
};

type HarEntry = {
    request: { method: string; url: string; bodySize?: number };
    response: { status: number; content?: { mimeType?: string; size?: number } };
};

export function harArtifactPaths(testInfo: TestInfo, rootDirectory: string): HarArtifactPaths {
    const relativeFile = path.relative(testInfo.config.rootDir, testInfo.file);
    const identity = JSON.stringify([
        relativeFile,
        testInfo.titlePath,
        testInfo.testId,
        testInfo.project.name,
        testInfo.repeatEachIndex,
    ]);
    const digest = createHash('sha256').update(identity).digest('hex').slice(0, 12);
    const label =
        [path.basename(relativeFile, path.extname(relativeFile)), ...testInfo.titlePath.slice(1), testInfo.project.name]
            .join('-')
            .normalize('NFKD')
            .replace(/[^a-zA-Z0-9_-]+/g, '-')
            .replace(/^-+|-+$/g, '')
            .slice(0, 90) || 'test';
    const directory = path.join(path.resolve(testInfo.config.rootDir, rootDirectory), `${label}-${digest}`);
    return {
        directory,
        recording: path.join(directory, 'recording.har'),
        requests: path.join(directory, 'requests.json'),
        responses: path.join(directory, 'responses.json'),
    };
}

export async function readHar(file: string): Promise<HarEntry[]> {
    let parsed: unknown;
    try {
        parsed = JSON.parse(await readFile(file, 'utf8'));
    } catch (error) {
        throw new Error(`Cannot read HAR at ${file}: ${error instanceof Error ? error.message : String(error)}`, {
            cause: error,
        });
    }
    const log = (parsed as { log?: { version?: unknown; entries?: unknown } })?.log;
    const entries = log?.entries;
    if (
        typeof log?.version !== 'string' ||
        !Array.isArray(entries) ||
        entries.some(
            (entry) =>
                !entry ||
                typeof entry.request?.method !== 'string' ||
                typeof entry.request?.url !== 'string' ||
                typeof entry.response?.status !== 'number'
        )
    ) {
        throw new Error(`Malformed HAR at ${file}: expected log.entries with request and response data`);
    }
    return entries as HarEntry[];
}

export async function readJsonArtifact<T>(file: string): Promise<T> {
    return JSON.parse(await readFile(file, 'utf8')) as T;
}

export async function writeJsonArtifact(file: string, value: unknown): Promise<void> {
    await mkdir(path.dirname(file), { recursive: true });
    const temporary = `${file}.${process.pid}.${createHash('sha256').update(String(Math.random())).digest('hex').slice(0, 8)}.tmp`;
    await writeFile(temporary, `${JSON.stringify(value, null, 2)}\n`, 'utf8');
    await rename(temporary, file);
}

export async function summarizeHar(file: string, paths: HarArtifactPaths): Promise<void> {
    const entries = await readHar(file);
    const requests: HarRequestSummary[] = entries.map(({ request }) => ({
        method: request.method,
        url: request.url,
        bodySize: typeof request.bodySize === 'number' ? request.bodySize : null,
    }));
    const responses: HarResponseSummary[] = entries.map(({ request, response }) => ({
        method: request.method,
        url: request.url,
        status: response.status,
        mimeType: response.content?.mimeType ?? '',
        bodySize: typeof response.content?.size === 'number' ? response.content.size : null,
    }));
    await writeJsonArtifact(paths.requests, requests);
    await writeJsonArtifact(paths.responses, responses);
}

export async function publishHar(temporary: string, paths: HarArtifactPaths): Promise<void> {
    // Validate before replacing the canonical recording. A failed close or malformed capture
    // therefore cannot overwrite a usable fixture.
    await readHar(temporary);
    await mkdir(paths.directory, { recursive: true });
    await rename(temporary, paths.recording);
    await summarizeHar(paths.recording, paths);
}
