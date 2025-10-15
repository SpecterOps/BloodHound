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

import react from '@vitejs/plugin-react';
import { resolve } from 'path';
import { defineConfig } from 'vite';
import dts from 'vite-plugin-dts';

export default defineConfig({
    plugins: [react(), dts({ exclude: ['**/*.stories.ts'] })],
    resolve: {
        alias: {
            components: resolve(__dirname, 'src/components'),
        },
    },
    build: {
        lib: {
            entry: resolve(__dirname, 'src/index.ts'),
            name: 'doodle-ui',
            // },
            // sourcemap: true,
            // rollupOptions: {
            //     output: {
            //         dir: 'dist',
            //         format: 'es',
            //     },
            //     plugins: [
            //         typescript({
            //             exclude: ['**/*.test.tsx', '**/*.stories.tsx'],
            //         }),
            //         terser(),
            //         del({ targets: 'dist/*' }),
            //     ],
            //     external: ['react', 'react-dom', 'react/jsx-runtime', 'tailwindcss'],
        },
    },
});
