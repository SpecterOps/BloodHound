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
import { findIconDefinition } from '@fortawesome/fontawesome-svg-core';
import { IconName } from '@fortawesome/free-solid-svg-icons';
import { AssetGroupTag, AssetGroupTagTypeOwned } from 'js-client-library';
import { useEffect, useState } from 'react';
import {
    DEFAULT_GLYPH_BACKGROUND_COLOR,
    DEFAULT_GLYPH_COLOR,
    GLYPH_SCALE,
    getModifiedSvgUrlFromIcon,
} from '../../utils';
import { useTagsQuery } from './useAssetGroupTags';

const glyphQualifier = (glyph: string | null) => !glyph?.includes('http');

const glyphTransformer = (glyph: string, darkMode?: boolean): string => {
    const iconDefiniton = findIconDefinition({ prefix: 'fas', iconName: glyph as IconName });

    if (!iconDefiniton) return '';

    const glyphIconUrl = getModifiedSvgUrlFromIcon(iconDefiniton, {
        styles: {
            'transform-origin': 'center',
            scale: GLYPH_SCALE,
            background: darkMode ? DEFAULT_GLYPH_COLOR : DEFAULT_GLYPH_BACKGROUND_COLOR,
            color: darkMode ? DEFAULT_GLYPH_BACKGROUND_COLOR : DEFAULT_GLYPH_COLOR,
        },
    });

    return glyphIconUrl;
};

export interface GlyphUtils {
    qualifier?: (glyph: string | null) => boolean;
    transformer: (glyph: string, darkMode?: boolean) => string;
}

export const glyphUtils: GlyphUtils = {
    qualifier: glyphQualifier,
    transformer: glyphTransformer,
};

// The word "Label" here is used in the sense of a cypher Label,
// e.g., in the cypher query: `match(u:User) return u`, 'User' is a cypher Label.
// That is to say the "Label" usage here does not reflect a type of AssetGroupTag
export const TagLabelPrefix = 'Tag_' as const;

export const createGlyphMapFromTags = (
    tags: AssetGroupTag[] | undefined,
    utils: GlyphUtils,
    darkMode?: boolean
): Record<string, string> => {
    const glyphMap: Record<string, string> = {};
    const { qualifier = () => true, transformer } = utils;

    tags?.forEach((tag) => {
        const underscoredTagName = tag.name.split(' ').join('_');

        if (tag.glyph === null) return;
        if (!qualifier(tag.glyph)) return;

        const glyphValue = transformer(tag.glyph, darkMode);

        if (tag.type === AssetGroupTagTypeOwned) {
            glyphMap.owned = `${TagLabelPrefix}${underscoredTagName}`;
            glyphMap.ownedGlyph = glyphValue;
            return;
        }

        if (glyphValue !== '') glyphMap[`${TagLabelPrefix}${underscoredTagName}`] = glyphValue;
    });

    return glyphMap;
};

export const getGlyphFromKinds = (kinds: string[] = [], tagGlyphMap: Record<string, string> = {}): string | null => {
    for (let index = kinds.length - 1; index > -1; index--) {
        const kind = kinds[index];

        if (!kind.includes(TagLabelPrefix)) continue;

        if (tagGlyphMap[kind]) return tagGlyphMap[kind];
    }
    return null;
};

export type TagGlyphs = Record<string, string>;

export const useTagGlyphs = (glyphUtils: GlyphUtils, darkMode?: boolean): TagGlyphs => {
    const [glyphMap, setGlyphMap] = useState<Record<string, string>>({});
    const tagsQuery = useTagsQuery();

    useEffect(() => {
        if (!tagsQuery.data) return;

        const newMap = createGlyphMapFromTags(tagsQuery.data, glyphUtils, darkMode);
        setGlyphMap(newMap);
    }, [tagsQuery.data, glyphUtils, darkMode]);

    return glyphMap;
};
