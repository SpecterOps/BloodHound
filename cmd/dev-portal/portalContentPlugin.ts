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
import { glob } from 'tinyglobby';
import type { Plugin, ViteDevServer } from 'vite';
import { parsePortalConfig, type PortalConfig } from './src/config/portalConfig';
import { resolveRepoPath } from './src/docs/links';

export const PORTAL_CONTENT_MODULE_ID = 'virtual:dev-portal';
const RESOLVED_MODULE_ID = '\0' + PORTAL_CONTENT_MODULE_ID;

const MARKDOWN_IMAGE_PATTERN = /!\[[^\]]*\]\(\s*<?([^)\s>]+)/g;
const HTML_IMAGE_PATTERN = /<img\b[^>]*?\bsrc\s*=\s*["']([^"']+)["']/gi;
const EXTERNAL_URL_PATTERN = /^([a-z][a-z0-9+.-]*:|\/\/)/i;

export type PortalContent = {
    config: PortalConfig;
    docs: { path: string; content: string }[];
    // Repo-relative image path to absolute file path, for images referenced by the docs
    images: Map<string, string>;
};

const toPosix = (filePath: string) => filePath.split(path.sep).join('/');

const findImageReferences = (content: string) =>
    [...content.matchAll(MARKDOWN_IMAGE_PATTERN), ...content.matchAll(HTML_IMAGE_PATTERN)].map((match) => match[1]);

// Reads conf.json and gathers every doc and doc image it selects. Throws when the config is invalid
// or the default page is not among the selected docs; include patterns that match nothing only warn.
export const collectPortalContent = async (
    configPath: string,
    repoRoot: string,
    warn: (message: string) => void
): Promise<PortalContent> => {
    const config = parsePortalConfig(JSON.parse(fs.readFileSync(configPath, 'utf8')));
    const docPaths = new Set<string>();

    for (const pattern of config.docs.include) {
        const matches = await glob(pattern, { cwd: repoRoot, ignore: config.docs.exclude, onlyFiles: true });
        if (matches.length === 0) warn(`Dev Portal docs pattern "${pattern}" matched no files`);
        matches.forEach((match) => docPaths.add(toPosix(match)));
    }

    if (!docPaths.has(config.docs.defaultPage)) {
        throw new Error(`Dev Portal docs.defaultPage "${config.docs.defaultPage}" is not among the included docs`);
    }

    const images = new Map<string, string>();
    const docs = [...docPaths].sort().map((docPath) => {
        const content = fs.readFileSync(path.join(repoRoot, docPath), 'utf8');
        findImageReferences(content)
            .filter((reference) => !EXTERNAL_URL_PATTERN.test(reference))
            .forEach((reference) => {
                const imagePath = resolveRepoPath(docPath, reference.split(/[?#]/)[0]);
                const absolutePath = imagePath && path.join(repoRoot, imagePath);
                if (absolutePath && fs.existsSync(absolutePath) && fs.statSync(absolutePath).isFile()) {
                    images.set(imagePath, absolutePath);
                }
            });
        return { path: docPath, content };
    });

    return { config, docs, images };
};

export const generatePortalModule = ({ config, docs, images }: PortalContent) => {
    const imageEntries = [...images.entries()];
    const imports = imageEntries.map(
        ([, absolutePath], index) => `import image${index} from ${JSON.stringify(toPosix(absolutePath) + '?url')};`
    );
    const imageMap = imageEntries.map(([imagePath], index) => `${JSON.stringify(imagePath)}: image${index}`);
    const { include: _include, exclude: _exclude, ...docsConfig } = config.docs;

    return [
        ...imports,
        `export const cards = ${JSON.stringify(config.cards)};`,
        `export const docsConfig = ${JSON.stringify(docsConfig)};`,
        `export const docs = ${JSON.stringify(docs)};`,
        `export const images = {${imageMap.join(', ')}};`,
    ].join('\n');
};

// Serves conf.json's cards and selected docs as the `virtual:dev-portal` module. Under the dev server,
// editing conf.json or any selected doc reloads the page.
export const portalContentPlugin = ({ configPath, repoRoot }: { configPath: string; repoRoot: string }): Plugin => {
    let devServer: ViteDevServer | undefined;
    let watchedFiles = new Set<string>([configPath]);

    return {
        name: 'dev-portal-content',
        configureServer(server) {
            devServer = server;
            server.watcher.add(configPath);
        },
        resolveId(id) {
            return id === PORTAL_CONTENT_MODULE_ID ? RESOLVED_MODULE_ID : undefined;
        },
        async load(id) {
            if (id !== RESOLVED_MODULE_ID) return;
            try {
                const content = await collectPortalContent(configPath, repoRoot, (message) => this.warn(message));
                watchedFiles = new Set([configPath, ...content.docs.map((doc) => path.join(repoRoot, doc.path))]);
                watchedFiles.forEach((file) => this.addWatchFile(file));
                devServer?.watcher.add([...watchedFiles]);
                return generatePortalModule(content);
            } catch (error) {
                this.error(error instanceof Error ? error.message : String(error));
            }
        },
        handleHotUpdate({ file, server }) {
            if (!watchedFiles.has(path.normalize(file))) return;
            const portalModule = server.moduleGraph.getModuleById(RESOLVED_MODULE_ID);
            if (portalModule) server.moduleGraph.invalidateModule(portalModule);
            server.ws.send({ type: 'full-reload' });
            return [];
        },
    };
};
