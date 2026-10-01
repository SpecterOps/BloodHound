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

import type { ThemeCatalogEntry } from './types';

export const themeCatalog: readonly ThemeCatalogEntry[] = [
    {
        id: 'retro-windows',
        name: 'Retro Windows',
        author: 'BloodHound Hackathon',
        description: 'Classic desktop colors with optional paperclip assistant and cursor.',
        packagePath: 'themes/retro-windows.json',
        preview: ['#C0C0C0', '#000080', '#008080'],
    },
    {
        id: 'synthwave',
        name: 'Synthwave',
        author: 'BloodHound Hackathon',
        description: 'Electric violet and cyan colors for a late-night graph session.',
        packagePath: 'themes/synthwave.json',
        preview: ['#211538', '#BA68FF', '#30D5C8'],
    },
    {
        id: 'wingdings',
        name: 'Wingdings',
        author: 'BloodHound Hackathon',
        description: 'High-contrast ink and paper colors with Wingdings lettering where installed.',
        packagePath: 'themes/wingdings.json',
        preview: ['#FAFAF7', '#222222', '#15537A'],
    },
];
