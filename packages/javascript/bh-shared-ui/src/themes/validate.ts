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

import type { ThemeColors, ThemeEffect, ThemePackage } from './types';

const colorPattern = /^#[0-9a-fA-F]{6}$/;
const effectNames: readonly ThemeEffect[] = ['clippy', 'retro-cursor'];

const isRecord = (value: unknown): value is Record<string, unknown> =>
    typeof value === 'object' && value !== null && !Array.isArray(value);

const isColor = (value: unknown): value is string => typeof value === 'string' && colorPattern.test(value);

const isThemeColors = (value: unknown): value is ThemeColors => {
    if (!isRecord(value)) return false;

    const colorKeys = [
        'primary',
        'primaryVariant',
        'secondary',
        'secondaryVariant',
        'onPrimary',
        'onSecondary',
        'text',
        'mutedText',
        'contrast',
        'link',
    ];

    return (
        colorKeys.every((key) => isColor(value[key])) &&
        Array.isArray(value.neutral) &&
        value.neutral.length === 5 &&
        value.neutral.every(isColor)
    );
};

export const parseThemePackage = (value: unknown, expectedId: string): ThemePackage => {
    if (!isRecord(value) || !isRecord(value.variants)) throw new Error('Invalid theme package');

    const effects = value.availableEffects;
    const validEffects =
        effects === undefined ||
        (Array.isArray(effects) && effects.every((effect) => effectNames.includes(effect as ThemeEffect)));

    if (
        value.schemaVersion !== 1 ||
        value.id !== expectedId ||
        typeof value.version !== 'string' ||
        typeof value.name !== 'string' ||
        typeof value.author !== 'string' ||
        typeof value.description !== 'string' ||
        (value.font !== undefined && value.font !== 'wingdings') ||
        !isThemeColors(value.variants.light) ||
        !isThemeColors(value.variants.dark) ||
        !validEffects
    ) {
        throw new Error('Invalid theme package');
    }

    return value as ThemePackage;
};
