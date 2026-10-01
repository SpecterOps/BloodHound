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
import type { StorybookConfig } from '@storybook/react-vite';
import autoprefixer from 'autoprefixer';
import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import tailwindcss from 'tailwindcss';
import tailwindConfig from './tailwind.config';

const require = createRequire(import.meta.url);

const getAbsolutePath = (packageName: string) => dirname(require.resolve(join(packageName, 'package.json')));

const config: StorybookConfig = {
    stories: ['../src/**/*.stories.@(ts|tsx)'],
    addons: [
        '@storybook/addon-links',
        '@storybook/addon-essentials',
        '@storybook/addon-interactions',
        '@storybook/addon-a11y',
        '@storybook/addon-themes',
    ].map(getAbsolutePath),
    framework: {
        name: getAbsolutePath('@storybook/react-vite'),
        options: {},
    },
    docs: { autodocs: 'tag' },
    async viteFinal(config) {
        const { mergeConfig } = await import('vite');
        return mergeConfig(config, {
            resolve: {
                alias: {
                    'doodle-ui': fileURLToPath(new URL('../../doodle-ui/src', import.meta.url)),
                    'js-client-library': fileURLToPath(new URL('../../js-client-library/src', import.meta.url)),
                },
            },
            css: {
                postcss: { plugins: [tailwindcss(tailwindConfig), autoprefixer()] },
            },
        });
    },
};

export default config;
