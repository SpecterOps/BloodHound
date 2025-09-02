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

import {
    DEFAULT_GLYPH_COLOR,
    DEFAULT_ICON_COLOR,
    GLYPHS,
    GLYPH_SCALE,
    GetIconInfo,
    GlyphDictionary,
    IconDictionary,
    NODE_ICONS,
    NODE_SCALE,
    UNKNOWN_ICON,
    getModifiedSvgUrlFromIcon,
} from 'bh-shared-ui';

// Adds object URLs to all icon definitions so that our fontawesome icons can be used by sigmajs node programs
const appendIconSvgUrls = (icons: IconDictionary): void => {
    Object.entries(icons).forEach(([type, value]) => {
        if (value.url) return;

        icons[type].url = getModifiedSvgUrlFromIcon(value.icon, {
            styles: { color: DEFAULT_ICON_COLOR, scale: NODE_SCALE, 'background-color': value.color },
        });
    });
};

// Adds object URLs to all glyph definitions so that our fontawesome icons can be used by sigmajs node programs
const appendGlyphSvgUrls = (icons: GlyphDictionary): void => {
    Object.entries(icons).forEach(([type, value]) => {
        if (value.url) return;

        icons[type].url = getModifiedSvgUrlFromIcon(value.icon, {
            styles: {
                color: value.iconColor ?? DEFAULT_GLYPH_COLOR,
                scale: GLYPH_SCALE,
            },
        });
    });
};

function transformIconDictionary(icons: IconDictionary): IconDictionary {
    appendIconSvgUrls(icons);

    return icons;
}

// Append URLs for nodes, glyphs, and any additional utility icons
appendIconSvgUrls(NODE_ICONS);
appendGlyphSvgUrls(GLYPHS);
UNKNOWN_ICON.url = getModifiedSvgUrlFromIcon(UNKNOWN_ICON.icon, { styles: { scale: NODE_SCALE } });

export { GLYPHS, GetIconInfo, NODE_ICONS, UNKNOWN_ICON, transformIconDictionary };
