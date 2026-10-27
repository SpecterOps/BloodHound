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

import { parsePortalConfig } from './portalConfig';

const validConfig = () => ({
    cards: [{ accent: 'blue', title: 'Docs', description: 'Read the docs.', href: '#/docs', icon: 'documentation' }],
    docs: {
        defaultPage: 'README.md',
        repositoryUrl: 'https://github.com/SpecterOps/BloodHound/blob/main/',
        include: ['**/*.md'],
        exclude: ['**/node_modules/**'],
    },
});

describe('parsePortalConfig', () => {
    it('accepts a valid config', () => {
        const config = validConfig();

        expect(parsePortalConfig(config)).toEqual(config);
    });

    it('reports every problem at once', () => {
        const config = validConfig();
        const invalid = {
            ...config,
            cards: [{ ...config.cards[0], accent: 'green', icon: 'rocket', title: '' }],
            docs: { ...config.docs, include: [], repositoryUrl: 'github.com/repo' },
        };

        expect(() => parsePortalConfig(invalid)).toThrow(
            [
                'Invalid Dev Portal config:',
                '  - cards[0].title must be a non-empty string',
                '  - cards[0].accent must be one of: blue, orange',
                '  - cards[0].icon must be one of: storybook, documentation',
                '  - docs.repositoryUrl must be an http(s) URL ending in "/"',
                '  - docs.include must be a non-empty list of glob patterns',
            ].join('\n')
        );
    });

    it('rejects unknown settings so typos are caught', () => {
        const config = { ...validConfig(), theme: 'dark' };

        expect(() => parsePortalConfig(config)).toThrow('config.theme is not a known setting');
    });

    it.each([
        ['a non-object config', []],
        ['missing sections', {}],
    ])('rejects %s', (_description, raw) => {
        expect(() => parsePortalConfig(raw)).toThrow('Invalid Dev Portal config');
    });
});
