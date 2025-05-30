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
        '@neo4j-cypher/react-codemirror',
        '@neo4j-cypher/codemirror/css/cypher-codemirror.css',
        '@emotion/react',
        '@emotion/styled',
        '@faker-js/faker',
        '@fortawesome/fontawesome-svg-core',
        '@fortawesome/free-solid-svg-icons',
        '@fortawesome/react-fontawesome',
        '@mui/material',
        '@mui/styles',
        '@mui/styles/makeStyles',
        '@mui/styles/withStyles',
        '@reduxjs/toolkit',
        'clsx',
        'downshift',
        'history',
        'immer',
        'immutable',
        'js-client-library',
        'js-file-download',
        'lodash/capitalize',
        'lodash/cloneDeep',
        'lodash/find',
        'lodash/isEmpty',
        'lodash/orderBy',
        'lodash/startCase',
        'lodash/toString',
        'luxon',
        'memoize-one',
        'msw',
        'notistack',
        'prop-types',
        'react',
        'react-error-boundary',
        'react-hook-form',
        'react-immutable-proptypes',
        'react-query',
        'react-router-dom',
        'react-window',
        'react-window-infinite-loader',
        'react/jsx-runtime',
        'swagger-ui-react',
        'swagger-ui-react/swagger-ui.css',
        'tailwind-merge',
        'tailwindcss',
    ],
};
