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

import terser from '@rollup/plugin-terser';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';
import del from 'rollup-plugin-delete';
import { defineConfig } from 'vite';
import dts from 'vite-plugin-dts';
import packageJson from './package.json';

export default defineConfig({
    plugins: [react(), dts({ exclude: ['**/*.stories.ts'], pathsToAliases: true })],
    resolve: {
        alias: {
            components: resolve(__dirname, 'src/components'),
            'components/*': resolve(__dirname, 'src/components/*'),
        },
    },
    build: {
        lib: {
            entry: resolve(__dirname, 'src'),
            formats: ['es'],
        },
        sourcemap: true,
        rollupOptions: {
            external: ['react', 'react-dom', 'react/jsx-runtime', 'tailwindcss'],
            output: {
                manualChunks: {
                    vendor: Object.keys(packageJson.dependencies),
                },
            },
            plugins: [del({ targets: 'dist/*' }), terser()],
        },
    },
});
