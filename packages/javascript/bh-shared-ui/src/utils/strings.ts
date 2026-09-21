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

import { type Nullable } from './type';

/** Represents a string literal type that can be widened to string, keeping intellisense for literal values */
// eslint-disable-next-line @typescript-eslint/ban-types
export type LiteralUnion<T extends string> = T | (string & {});

// Allows to truncate text and add an ellipsis
export const truncateText = (
    text: string | undefined,
    maxChars: number = 20,
    trailingChars: string = '...'
): string | undefined => {
    if (!text) return;
    if (text.length <= maxChars) return text;
    return text.slice(0, maxChars) + trailingChars;
};

export type KeywordAndTypeValues = {
    keyword: string | undefined;
    type: string | undefined;
};

export function parseKeywordAndTypeValue(
    inputValue: Nullable<string>,
    kinds: string[] | undefined
): KeywordAndTypeValues {
    const value = inputValue ?? undefined;

    if (!value) {
        return {
            keyword: undefined,
            type: undefined,
        };
    }

    const hasParsableQualifier = value.length > 1 && value.includes(':');

    if (!hasParsableQualifier) {
        return {
            keyword: value,
            type: undefined,
        };
    }

    const [qualifier, ...keywordParts] = value.split(':');
    const nodeKind = kinds?.find((kind) => kind.toLocaleLowerCase() === qualifier.toLocaleLowerCase());

    if (!nodeKind) {
        return {
            keyword: value,
            type: undefined,
        };
    }

    return {
        keyword: keywordParts.join(':'),
        type: nodeKind,
    };
}

/**
 * Parses a JSON object without throwing for invalid or unsupported JSON values.
 * Useful for validating JSON strings
 *
 * @param text - The JSON text to parse.
 * @returns The parsed object, a descriptive error message for invalid JSON, arrays, or nested properties, or
 * `undefined` for other unsupported JSON values.
 */
export const safeParseJson = (text: string, flat: boolean = true): Record<string, unknown> | string | undefined => {
    try {
        const params: unknown = JSON.parse(text);
        if (Array.isArray(params)) {
            return 'Use a JSON object as the top-level value. Arrays are not allowed.';
        }

        if (typeof params === 'object' && params !== null && !Array.isArray(params)) {
            if (flat) {
                if (Object.values(params).some((value) => typeof value === 'object' && value !== null)) {
                    return 'Must be a flat JSON object. Nested values are not supported.';
                }
            }

            return params as Record<string, unknown>;
        }
    } catch {
        return 'Syntax error: invalid JSON.';
    }

    return undefined;
};
