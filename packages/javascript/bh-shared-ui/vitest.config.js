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

import path from 'path';
import { defineConfig } from 'vitest/config';

export default defineConfig({
    resolve: {
        alias: {
            'js-client-library': path.resolve(__dirname, '..', 'js-client-library', 'src'),
        },
    },
    test: {
        globals: true,
        environment: 'jsdom',
        setupFiles: ['./src/setupTests.tsx'],
        testTimeout: 60000, // 1 minute,
        coverage: {
            provider: 'v8',
            reportsDirectory: './coverage',
            reporter: ['text', 'json', 'json-summary', 'html'],
            exclude: ['**/types/**', '**/constants/**', 'dist', '**/components/HelpTexts/**'],
            thresholds: {
                lines: 60,
                functions: 60,
                branches: 60,
                statements: 60,
            },
        },
        reporters: [
            'default',
            'github-actions',
            [
                'allure-vitest/reporter',
                {
                    resultsDir: '../../../allure-results',
                },
            ],
        ],
        enabled: true,
        reportOnFailure: true, // report coverage even if fails
    },
});
