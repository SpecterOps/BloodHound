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

import { getContrastRatio } from '@mui/material/styles';
import { darkPalette, lightPalette } from '../constants';
import { applyThemeToDocument, paletteForTheme } from './applyTheme';
import type { ThemeColors, ThemePackage } from './types';
import { parseThemePackage } from './validate';

const colors = (primary: string, background: string): ThemeColors => ({
    primary,
    primaryVariant: '#123456',
    secondary: '#654321',
    secondaryVariant: '#234567',
    onPrimary: '#FFFFFF',
    onSecondary: '#000000',
    text: '#F0F0F0',
    mutedText: '#AAAAAA',
    contrast: '#FFFFFF',
    link: '#789ABC',
    neutral: [background, '#222222', '#333333', '#444444', '#555555'],
});

const theme: ThemePackage = {
    schemaVersion: 1,
    id: 'retro-windows',
    version: '1.0.0',
    name: 'Retro Windows',
    author: 'BloodHound Hackathon',
    description: 'A theme',
    variants: {
        light: colors('#000080', '#C0C0C0'),
        dark: colors('#8189EB', '#1C1C24'),
    },
    availableEffects: ['clippy', 'retro-cursor'],
};

afterEach(() => applyThemeToDocument(undefined, 'light', []));

describe('theme packages', () => {
    it('validates IDs and color tokens before installing a package', () => {
        expect(parseThemePackage(theme, 'retro-windows')).toEqual(theme);
        expect(() => parseThemePackage(theme, 'synthwave')).toThrow('Invalid theme package');
        expect(() =>
            parseThemePackage({ ...theme, variants: { ...theme.variants, dark: colors('red', '#1C1C24') } }, theme.id)
        ).toThrow('Invalid theme package');
    });

    it('switches light and dark tokens, then restores defaults', () => {
        applyThemeToDocument(theme, 'light', ['retro-cursor']);
        expect(document.documentElement.style.getPropertyValue('--primary')).toBe('#000080');
        expect(document.documentElement.style.getPropertyValue('--neutral-1')).toBe('#C0C0C0');
        expect(document.documentElement).toHaveClass('theme-cursor');

        applyThemeToDocument(theme, 'dark', []);
        expect(document.documentElement.style.getPropertyValue('--primary')).toBe('#8189EB');
        expect(document.documentElement.style.getPropertyValue('--neutral-1')).toBe('#1C1C24');
        expect(document.documentElement.style.getPropertyValue('--neutral-light-1')).toBe('#C0C0C0');
        expect(document.documentElement.style.getPropertyValue('--neutral-dark-1')).toBe('#1C1C24');
        expect(document.documentElement).not.toHaveClass('theme-cursor');

        applyThemeToDocument(undefined, 'dark', []);
        expect(document.documentElement.style.getPropertyValue('--primary')).toBe('');
        expect(document.documentElement).not.toHaveClass('theme-installed');
    });

    it('leaves semantic risk colors unchanged in the MUI palette', () => {
        const light = paletteForTheme(theme, 'light');
        const dark = paletteForTheme(theme, 'dark');

        expect(light.primary.main).toBe(theme.variants.light.primary);
        expect(dark.primary.main).toBe(theme.variants.dark.primary);
        expect(light.critical).toBe(lightPalette.critical);
        expect(dark.high).toBe(darkPalette.high);
        expect(light.color.error).toBe(lightPalette.color.error);
    });

    it('keeps text readable against a primary variant from an already-installed package', () => {
        const installedTheme: ThemePackage = {
            ...theme,
            variants: {
                ...theme.variants,
                dark: { ...theme.variants.dark, primaryVariant: '#5C65C0', onPrimary: '#11111C' },
            },
        };

        applyThemeToDocument(installedTheme, 'dark', []);
        const cssVariant = document.documentElement.style.getPropertyValue('--primary-variant');
        const muiVariant = paletteForTheme(installedTheme, 'dark').primary.dark;

        expect(getContrastRatio(cssVariant, installedTheme.variants.dark.onPrimary)).toBeGreaterThanOrEqual(4.5);
        expect(cssVariant).toBe(muiVariant);
    });
});
