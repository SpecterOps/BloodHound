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
import os from 'os';
import path from 'path';
import { collectPortalContent, generatePortalModule } from './portalContentPlugin';

const writeFile = (root: string, relativePath: string, content: string) => {
    const filePath = path.join(root, relativePath);
    fs.mkdirSync(path.dirname(filePath), { recursive: true });
    fs.writeFileSync(filePath, content);
};

const docsConfig = (overrides: Record<string, unknown> = {}) => ({
    defaultPage: 'README.md',
    repositoryUrl: 'https://github.com/SpecterOps/BloodHound/blob/main/',
    include: ['**/*.md', '.github/**/*.md'],
    exclude: ['**/node_modules/**'],
    ...overrides,
});

describe('collectPortalContent', () => {
    let repoRoot: string;
    let configPath: string;

    const writeConfig = (docs: Record<string, unknown>) =>
        fs.writeFileSync(configPath, JSON.stringify({ cards: [], docs }));

    beforeEach(() => {
        repoRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'dev-portal-'));
        configPath = path.join(repoRoot, 'conf.json');
        writeFile(repoRoot, 'README.md', '![Logo](img/logo.svg)\n<img src="img/missing.png" alt="gone">');
        writeFile(repoRoot, 'img/logo.svg', '<svg />');
        writeFile(repoRoot, 'rfc/bh-rfc-0.md', '# RFC 0\n![Remote](https://example.com/a.png)');
        writeFile(repoRoot, '.github/pull_request_template.md', '# PR');
        writeFile(repoRoot, '.augment/notes.md', '# Private notes');
        writeFile(repoRoot, 'web/node_modules/pkg/README.md', '# Dependency');
    });

    afterEach(() => {
        fs.rmSync(repoRoot, { recursive: true, force: true });
    });

    it('selects included docs, skipping excluded and unlisted dot folders', async () => {
        writeConfig(docsConfig());

        const content = await collectPortalContent(configPath, repoRoot, vi.fn());

        expect(content.docs.map((doc) => doc.path)).toEqual([
            '.github/pull_request_template.md',
            'README.md',
            'rfc/bh-rfc-0.md',
        ]);
    });

    it('collects only repo images that exist', async () => {
        writeConfig(docsConfig());

        const { images } = await collectPortalContent(configPath, repoRoot, vi.fn());

        expect([...images.entries()]).toEqual([['img/logo.svg', path.join(repoRoot, 'img/logo.svg')]]);
    });

    it('warns about include patterns that match nothing', async () => {
        const warn = vi.fn();
        writeConfig(docsConfig({ include: ['**/*.md', 'missing/**/*.md'] }));

        await collectPortalContent(configPath, repoRoot, warn);

        expect(warn).toHaveBeenCalledWith('Dev Portal docs pattern "missing/**/*.md" matched no files');
    });

    it('fails when the default page is not an included doc', async () => {
        writeConfig(docsConfig({ defaultPage: 'MISSING.md' }));

        await expect(collectPortalContent(configPath, repoRoot, vi.fn())).rejects.toThrow(
            'Dev Portal docs.defaultPage "MISSING.md" is not among the included docs'
        );
    });

    it('fails on an invalid config', async () => {
        writeConfig(docsConfig({ include: 'README.md' }));

        await expect(collectPortalContent(configPath, repoRoot, vi.fn())).rejects.toThrow(
            'docs.include must be a non-empty list of glob patterns'
        );
    });

    it('generates a module exposing cards, docs config, docs, and image imports', async () => {
        writeConfig(docsConfig());

        const moduleSource = generatePortalModule(await collectPortalContent(configPath, repoRoot, vi.fn()));

        expect(moduleSource).toContain(`import image0 from ${JSON.stringify(`${repoRoot}/img/logo.svg?url`)};`);
        expect(moduleSource).toContain('export const cards = [];');
        expect(moduleSource).toContain(
            'export const docsConfig = {"defaultPage":"README.md","repositoryUrl":"https://github.com/SpecterOps/BloodHound/blob/main/"};'
        );
        expect(moduleSource).toContain('export const images = {"img/logo.svg": image0};');
    });
});
