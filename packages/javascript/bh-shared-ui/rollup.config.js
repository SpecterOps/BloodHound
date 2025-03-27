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
import typescript from '@rollup/plugin-typescript';
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
        '@bloodhoundenterprise/doodleui',
        '@emotion/react',
        '@emotion/styled',
        '@faker-js/faker',
        '@fortawesome/free-solid-svg-icons',
        '@fortawesome/fontawesome-svg-core',
        '@fortawesome/react-fontawesome',
        '@reduxjs/toolkit',
        '@mui/material',
        '@mui/styles',
        '@mui/styles/makeStyles',
        '@mui/styles/withStyles',
        'clsx',
        'tailwind-merge',
        'memoize-one',
        'react-error-boundary',
        'react-hook-form',
        'react-window-infinite-loader',
        'react-window',
        'react',
        'react/jsx-runtime',
        'luxon',
        'downshift',
        'notistack',
        'react-query',
        'js-client-library',
        'js-file-download',
        'swagger-ui-react',
        'swagger-ui-react/swagger-ui.css',
        'prop-types',
        'immutable',
        'immer',
        'msw',
        'react-immutable-proptypes',
        'lodash/toString',
        'lodash/cloneDeep',
        'react-router-dom',
        'lodash/isEmpty',
        'lodash/startCase',
        'lodash/find',
    ],
};
