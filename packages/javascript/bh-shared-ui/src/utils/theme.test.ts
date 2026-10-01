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

import { describe, expect, it } from 'vitest';
import { themePresets } from '../constants';
import { applyThemePresetCssVariables, getThemePresetCssVariables } from './theme';

describe('getThemePresetCssVariables', () => {
    it('only sets font variables for the default preset', () => {
        const variables = getThemePresetCssVariables(themePresets.default, false);

        expect(variables).toEqual({
            '--font-body': themePresets.default.bodyFontFamily,
            '--font-heading': themePresets.default.headingFontFamily,
        });
    });

    it('maps the active palette onto doodle-ui color variables for presets that override colors', () => {
        const preset = themePresets.comicSans;

        const lightVariables = getThemePresetCssVariables(preset, false);
        const darkVariables = getThemePresetCssVariables(preset, true);

        expect(lightVariables['--font-heading']).toBe(preset.headingFontFamily);
        expect(lightVariables['--primary']).toBe(preset.lightPalette.primary.main);
        expect(lightVariables['--neutral-1']).toBe(preset.lightPalette.neutral.primary);
        expect(darkVariables['--primary']).toBe(preset.darkPalette.primary.main);
        expect(darkVariables['--neutral-1']).toBe(preset.darkPalette.neutral.primary);
    });

    it('maps both palettes onto the fixed light and dark neutral variables', () => {
        const preset = themePresets.serif;

        const variables = getThemePresetCssVariables(preset, false);

        expect(variables['--neutral-light-5']).toBe(preset.lightPalette.neutral.quinary);
        expect(variables['--neutral-dark-5']).toBe(preset.darkPalette.neutral.quinary);
    });
});

describe('applyThemePresetCssVariables', () => {
    it('sets variables on the document root and removes stale ones when switching presets', () => {
        const rootStyle = window.document.documentElement.style;

        applyThemePresetCssVariables(themePresets.serif, false);
        expect(rootStyle.getPropertyValue('--font-heading')).toBe(themePresets.serif.headingFontFamily);
        expect(rootStyle.getPropertyValue('--primary')).toBe(themePresets.serif.lightPalette.primary.main);

        applyThemePresetCssVariables(themePresets.default, false);
        expect(rootStyle.getPropertyValue('--font-heading')).toBe(themePresets.default.headingFontFamily);
        expect(rootStyle.getPropertyValue('--primary')).toBe('');
    });
});
