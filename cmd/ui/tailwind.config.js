// Copyright 2025 Specter Ops, Inc.
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

import { DoodleUIPlugin, DoodleUIPreset } from '@bloodhoundenterprise/doodleui';

/** @type {import('tailwindcss').Config} */
export default {
    theme: {
        extend: {
            spacing: {
                'nav-width': '3.5rem',
                'subnav-width': '14rem',
                'nav-width-expanded': '17.5rem',
            },
            zIndex: {
                nav: '1200',
            },
        },
    },
    content: [
        './index.html',
        './src/**/*.{js,ts,jsx,tsx}',
        './node_modules/@bloodhoundenterprise/doodleui/dist/doodleui.js',
        './node_modules/bh-shared-ui/src/**/*.{js,ts,jsx,tsx}',
    ],
    darkMode: ['class'],
    plugins: [DoodleUIPlugin],
    presets: [DoodleUIPreset],
};
