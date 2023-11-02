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

import typescript from '@rollup/plugin-typescript';
import terser from '@rollup/plugin-terser';
import del from 'rollup-plugin-delete';

export default {
    input: 'src/index.ts',
    output: {
        dir: 'dist',
        format: 'esm',
        sourcemap: true,
    },
    plugins: [
        typescript({
            exclude: ['**/*.test.*', 'src/setupTests.tsx'],
        }),
        terser(),
        del({ targets: 'dist/*' }),
    ],
    external: [
        '@emotion/react',
        '@emotion/styled',
        '@fortawesome/free-solid-svg-icons',
        '@fortawesome/fontawesome-svg-core',
        '@fortawesome/react-fontawesome',
        '@reduxjs/toolkit',
        '@mui/material',
        '@mui/styles',
        '@mui/styles/makeStyles',
        '@mui/styles/withStyles',
        'clsx',
        'dompurify',
        'memoize-one',
        'react-hook-form',
        'react-markdown',
        'react-window-infinite-loader',
        'react-window',
        'react',
        'react/jsx-runtime',
        'luxon',
        'downshift',
        'notistack',
        'react-query',
        'js-client-library',
    ],
};
