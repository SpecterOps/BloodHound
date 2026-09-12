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

import { ActiveDirectoryNodeKind, DEFAULT_ICON_COLOR, IconDictionary, NODE_SCALE } from 'bh-shared-ui';

const getModifiedSvgUrlFromIconMock = vi.hoisted(() => vi.fn(() => 'blob:test-icon'));

vi.mock('bh-shared-ui', async (importOriginal) => ({
    ...(await importOriginal<typeof import('bh-shared-ui')>()),
    getModifiedSvgUrlFromIcon: getModifiedSvgUrlFromIconMock,
}));

import { NODE_ICONS, transformIconDictionary } from './svgIcons';

describe('Explore SVG icons', () => {
    it('creates transparent node icon images', () => {
        getModifiedSvgUrlFromIconMock.mockClear();
        const groupIcon = NODE_ICONS[ActiveDirectoryNodeKind.Group].icon;
        const icons: IconDictionary = { Group: { icon: groupIcon, color: '#DBE617' } };

        transformIconDictionary(icons);

        expect(getModifiedSvgUrlFromIconMock).toHaveBeenCalledOnce();
        expect(getModifiedSvgUrlFromIconMock).toHaveBeenCalledWith(groupIcon, {
            styles: { color: DEFAULT_ICON_COLOR, scale: NODE_SCALE },
        });
        expect(icons.Group.url).toBe('blob:test-icon');
    });
});
