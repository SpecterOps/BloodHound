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

/// <reference types="vitest/config" />
import react from '@vitejs/plugin-react';
import fs from 'fs';
import path from 'path';
import { defineConfig, searchForWorkspaceRoot, type Connect, type Plugin } from 'vite';
import { portalContentPlugin } from './portalContentPlugin';

const STORYBOOK_URL_PREFIX = '/doodle-ui-storybook/';
const STORYBOOK_BUILD_DIR = path.resolve(
    __dirname,
    '..',
    '..',
    'packages',
    'javascript',
    'doodle-ui',
    'storybook-static'
);

const CONTENT_TYPES: Record<string, string> = {
    '.css': 'text/css',
    '.html': 'text/html',
    '.js': 'text/javascript',
    '.json': 'application/json',
    '.map': 'application/json',
    '.mjs': 'text/javascript',
    '.png': 'image/png',
    '.svg': 'image/svg+xml',
    '.woff': 'font/woff',
    '.woff2': 'font/woff2',
};

// The Storybook page embeds the Doodle UI Storybook static build from /doodle-ui-storybook/. Under
// `just dev-portal`, nginx serves that path from a read-only mount (see hosting/nginx.conf), and the
// justfile rebuilds Storybook first when it is missing or older than the Doodle UI sources. The Vite
// dev and preview servers have no such path, so this plugin serves the same build directory there. Unlike
// `just dev-portal`, it never builds or checks freshness: it serves whatever is currently in
// storybook-static, and responds 404 until `yarn workspace doodle-ui build:storybook` has been run.
//
// Serving happens in a post hook so Vite's host check runs first. The requested path is captured
// beforehand because Vite's SPA fallback rewrites request.url to /index.html before post hooks run.
const storybookRequestPaths = new WeakMap<Connect.IncomingMessage, string>();

const captureDoodleUIStorybookPath: Connect.NextHandleFunction = (request, _response, next) => {
    const requestPath = decodeURIComponent((request.url ?? '').split('?')[0]);
    if (requestPath.startsWith(STORYBOOK_URL_PREFIX)) storybookRequestPaths.set(request, requestPath);
    next();
};

const doodleUIStorybookMiddleware: Connect.NextHandleFunction = (request, response, next) => {
    const requestPath = storybookRequestPaths.get(request);
    if (requestPath === undefined) return next();

    let filePath = path.join(STORYBOOK_BUILD_DIR, requestPath.slice(STORYBOOK_URL_PREFIX.length));
    if (filePath !== STORYBOOK_BUILD_DIR && !filePath.startsWith(STORYBOOK_BUILD_DIR + path.sep)) {
        response.statusCode = 403;
        return response.end();
    }
    if (fs.existsSync(filePath) && fs.statSync(filePath).isDirectory()) {
        filePath = path.join(filePath, 'index.html');
    }
    if (!fs.existsSync(filePath)) {
        response.statusCode = 404;
        return response.end('Doodle UI Storybook not found. Run `yarn workspace doodle-ui build:storybook`.');
    }

    response.setHeader('Content-Type', CONTENT_TYPES[path.extname(filePath)] ?? 'application/octet-stream');
    fs.createReadStream(filePath).pipe(response);
};

const serveDoodleUIStorybook = (): Plugin => ({
    name: 'serve-doodle-ui-storybook',
    configureServer(server) {
        server.middlewares.use(captureDoodleUIStorybookPath);
        return () => server.middlewares.use(doodleUIStorybookMiddleware);
    },
    configurePreviewServer(server) {
        server.middlewares.use(captureDoodleUIStorybookPath);
        return () => server.middlewares.use(doodleUIStorybookMiddleware);
    },
});

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        react(),
        serveDoodleUIStorybook(),
        portalContentPlugin({
            configPath: path.resolve(__dirname, 'conf.json'),
            repoRoot: path.resolve(__dirname, '..', '..'),
        }),
    ],
    resolve: {
        alias: {
            'doodle-ui': path.resolve(__dirname, '..', '..', 'packages', 'javascript', 'doodle-ui', 'src'),
        },
        dedupe: [
            'doodle-ui',
            '@fortawesome/fontawesome-svg-core',
            '@fortawesome/free-solid-svg-icons',
            '@fortawesome/react-fontawesome',
            'react',
            'react-dom',
            'react-router-dom',
            'tailwindcss',
        ],
        preserveSymlinks: true,
    },
    server: {
        port: 3010,
        fs: {
            allow: [searchForWorkspaceRoot(process.cwd())],
        },
    },
    test: {
        globals: true,
        environment: 'jsdom',
        setupFiles: ['./src/setupTests.ts'],
    },
    build: {
        outDir: './dist',
    },
});
