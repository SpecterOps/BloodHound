// Copyright 2023 Specter Ops, Inc.
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

/// <reference types="vitest" />
import react from '@vitejs/plugin-react';
import fs from 'fs';
import path from 'path';
import { defineConfig, loadEnv, Plugin, searchForWorkspaceRoot } from 'vite';
import glsl from 'vite-plugin-glsl';
import { configDefaults } from 'vitest/config';

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd(), '');

    return {
        plugins: [react(), glsl(), excludeMockServiceWorker(env)],
        resolve: {
            alias: {
                src: path.resolve(__dirname, './src'),
                'bh-shared-ui': path.resolve(__dirname, '..', '..', 'packages', 'javascript', 'bh-shared-ui', 'src'),
                'js-client-library': path.resolve(
                    __dirname,
                    '..',
                    '..',
                    'packages',
                    'javascript',
                    'js-client-library',
                    'src'
                ),
                'doodle-ui': path.resolve(__dirname, '..', '..', 'packages', 'javascript', 'doodle-ui', 'src'),
                'react-helmet-async': path.resolve(__dirname, '..', '..', 'node_modules', 'react-helmet-async'),
            },
            dedupe: [
                'doodle-ui',
                '@emotion/react',
                '@emotion/styled',
                '@faker-js/faker',
                '@fortawesome/fontawesome-free',
                '@fortawesome/fontawesome-svg-core',
                '@fortawesome/free-solid-svg-icons',
                '@fortawesome/react-fontawesome',
                '@mona-health/react-input-mask',
                '@mui/material',
                '@mui/styles',
                '@mui/lab',
                'downshift',
                'history',
                'notistack',
                'msw',
                'react',
                'react-error-boundary',
                'react-hook-form',
                'react-query',
                'react-helmet-async',
                'react-router-dom',
                'tailwindcss',
            ],
            preserveSymlinks: true,
        },
        base: '/ui',
        server: {
            proxy: {
                '/api': {
                    target: env.TARGET_PROXY_URL || 'http://localhost:8080',
                    changeOrigin: true,
                },
            },
            port: 3000,
            host: true,
            // When running behind the Traefik reverse proxy (docker compose dev), the browser reaches
            // Vite on the proxy port, not the container's internal 3000. VITE_HMR_CLIENT_PORT points the
            // HMR websocket back through the proxy. Left as `true` (default) for local `yarn dev`.
            hmr: env.VITE_HMR_CLIENT_PORT ? { clientPort: Number(env.VITE_HMR_CLIENT_PORT) } : true,
            // macOS Docker bind mounts don't reliably deliver fs events into the container, so enable
            // polling there (VITE_USE_POLLING=true) to detect source changes. Off locally to save CPU.
            // A larger interval and ignored heavy paths keep polling from starving Vite on slower Docker setups.
            watch:
                env.VITE_USE_POLLING === 'true'
                    ? {
                          usePolling: true,
                          interval: 300,
                          binaryInterval: 1000,
                          ignored: ['**/node_modules/**', '**/dist/**', '**/coverage/**', '**/.git/**'],
                      }
                    : undefined,
            fs: {
                allow: [searchForWorkspaceRoot(process.cwd())],
            },
        },
        preview: {
            port: 3000,
        },
        test: {
            globals: true,
            environment: 'jsdom',
            exclude: [
                ...configDefaults.exclude,
                // Playwright accessibility regression suite — run manually via `yarn test:a11y`.
                'tests/**',
            ],
            setupFiles: ['./src/setupTests.tsx'],
            testTimeout: 60000, // 1 minute,
            coverage: {
                provider: 'v8',
                reportsDirectory: './coverage',
                reporter: ['text-summary', 'json-summary'],
            },
            reporters: [
                'default',
                [
                    'allure-vitest/reporter',
                    {
                        resultsDir: '../../allure-results',
                    },
                ],
            ],
        },
        build: {
            outDir: env.BUILD_PATH || './dist',
        },
    };
});

/**
 * Exclude mockServiceWorker.js from production builds
 * MSW service worker should only be available in development mode
 */
function excludeMockServiceWorker(env: ReturnType<typeof loadEnv>): Plugin {
    return {
        name: 'exclude-mock-service-worker',
        apply: 'build', // Only apply during build (production)
        closeBundle() {
            // Remove mockServiceWorker.js from the output directory after build
            const mockServiceWorkerPath = env.BUILD_PATH
                ? path.resolve(env.BUILD_PATH, 'mockServiceWorker.js')
                : path.resolve(__dirname, 'dist', 'mockServiceWorker.js');

            if (fs.existsSync(mockServiceWorkerPath)) {
                fs.unlinkSync(mockServiceWorkerPath);
            }
        },
    };
}
