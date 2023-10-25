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

import { icon } from '@fortawesome/fontawesome-svg-core';
import { IconDefinition } from '@fortawesome/free-solid-svg-icons';
import { GLYPHS, GlyphDictionary, IconDictionary, NODE_ICON, UNKNOWN_ICON } from 'bh-shared-ui';

const NODE_SCALE = '0.6';
const GLYPH_SCALE = '0.5';
const DEFAULT_ICON_COLOR = '#000000';

// Adds object URLs to all icon and glyph definitions so that our fontawesome icons can be used by sigmajs node programs
const appendSvgUrls = (icons: IconDictionary | GlyphDictionary, scale: string): void => {
    Object.entries(icons).forEach(([type, value]) => {
        if (value.url) return;

        const color = (icons[type] as any).iconColor || DEFAULT_ICON_COLOR;
        icons[type].url = getModifiedSvgUrlFromIcon(value.icon, scale, color);
    });
};

const getModifiedSvgUrlFromIcon = (iconDefinition: IconDefinition, scale: string, color: string): string => {
    const modifiedIcon = icon(iconDefinition, {
        styles: { 'transform-origin': 'center', scale, color },
    });

    const svgString = modifiedIcon.html[0].replace(/<svg/, '<svg width="200" height="200"');
    const svg = new Blob([svgString], { type: 'image/svg+xml' });
    return URL.createObjectURL(svg);
};

// Append URLs for nodes, glyphs, and any additional utility icons
appendSvgUrls(NODE_ICON, NODE_SCALE);
appendSvgUrls(GLYPHS, GLYPH_SCALE);
UNKNOWN_ICON.url = getModifiedSvgUrlFromIcon(UNKNOWN_ICON.icon, NODE_SCALE, DEFAULT_ICON_COLOR);

export { NODE_ICON, GLYPHS, UNKNOWN_ICON };
