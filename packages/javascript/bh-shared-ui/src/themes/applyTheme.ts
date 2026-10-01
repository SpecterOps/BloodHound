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

import { darken, getContrastRatio, getLuminance, lighten } from '@mui/material/styles';
import createPalette, { Palette } from '@mui/material/styles/createPalette';
import { darkPalette, lightPalette } from '../constants';
import type { ThemeColors, ThemeEffect, ThemeMode, ThemePackage } from './types';

const readableVariant = (background: string, foreground: string): string => {
    if (getContrastRatio(background, foreground) >= 4.5) return background;

    const adjust = getLuminance(foreground) > 0.5 ? darken : lighten;
    for (let step = 1; step <= 20; step++) {
        const candidate = adjust(background, step / 20);
        if (getContrastRatio(candidate, foreground) >= 4.5) return candidate;
    }

    return adjust(background, 1);
};

const cssVariablesFor = (
    colors: ThemeColors,
    lightColors: ThemeColors,
    darkColors: ThemeColors
): Record<string, string> => {
    const [background, surface, raised, hover, border] = colors.neutral;
    const [lightBackground, lightSurface, lightRaised, lightHover, lightBorder] = lightColors.neutral;
    const [darkBackground, darkSurface, darkRaised, darkHover, darkBorder] = darkColors.neutral;

    return {
        '--primary': colors.primary,
        '--primary-main': colors.primary,
        '--primary-variant': readableVariant(colors.primaryVariant, colors.onPrimary),
        '--secondary': colors.secondary,
        '--secondary-main': colors.secondary,
        '--secondary-variant': colors.secondaryVariant,
        '--text-main': colors.text,
        '--text-light': colors.mutedText,
        '--text-muted': colors.mutedText,
        '--text-contrast': colors.onPrimary,
        '--contrast': colors.contrast,
        '--link': colors.link,
        '--link-main': colors.link,
        '--link-hover': colors.secondaryVariant,
        '--focus-ring': colors.secondary,
        '--focus-ring-offset': background,
        '--neutral-1': background,
        '--neutral-2': surface,
        '--neutral-3': raised,
        '--neutral-4': hover,
        '--neutral-5': border,
        '--neutral-50': background,
        '--neutral-100': background,
        '--neutral-200': surface,
        '--neutral-300': surface,
        '--neutral-400': raised,
        '--neutral-500': raised,
        '--neutral-600': hover,
        '--neutral-700': border,
        '--neutral-800': border,
        '--neutral-900': border,
        '--neutral-light-1': lightBackground,
        '--neutral-light-2': lightSurface,
        '--neutral-light-3': lightRaised,
        '--neutral-light-4': lightHover,
        '--neutral-light-5': lightBorder,
        '--neutral-dark-1': darkBackground,
        '--neutral-dark-2': darkSurface,
        '--neutral-dark-3': darkRaised,
        '--neutral-dark-4': darkHover,
        '--neutral-dark-5': darkBorder,
        '--elevation-1': background,
        '--elevation-2': surface,
        '--elevation-3': raised,
        '--elevation-4': hover,
        '--elevation-5': border,
        '--secondary-btn-fill': raised,
        '--secondary-btn-active-fill': hover,
        '--transparent-btn-border': border,
        '--toggle-btn-fill': surface,
        '--toggle-btn-border': border,
        '--toggle-group-fill': raised,
        '--input-fill': surface,
        '--input-border-default': border,
        '--input-outlined-border-default': border,
        '--input-outlined-border-hover': colors.secondary,
        '--textarea-fill': surface,
        '--textarea-border-default': border,
        '--select-trigger-fill': surface,
        '--select-trigger-outlined-fill': surface,
        '--select-content-fill': raised,
        '--select-border-default': border,
        '--dropdown-popover-fill': raised,
        '--dropdown-popover-border': border,
        '--dropdown-option-hover-fill': hover,
        '--data-table-fill': surface,
        '--data-table-header-fill': raised,
        '--data-table-row-even-fill': surface,
        '--data-table-row-odd-fill': background,
        '--data-table-row-hover-fill': hover,
    };
};

const emptyColors: ThemeColors = {
    primary: '#000000',
    primaryVariant: '#000000',
    secondary: '#000000',
    secondaryVariant: '#000000',
    onPrimary: '#000000',
    onSecondary: '#000000',
    text: '#000000',
    mutedText: '#000000',
    contrast: '#000000',
    link: '#000000',
    neutral: ['#000000', '#000000', '#000000', '#000000', '#000000'],
};
const themeVariableNames = Object.keys(cssVariablesFor(emptyColors, emptyColors, emptyColors));

export const applyThemeToDocument = (theme: ThemePackage | undefined, mode: ThemeMode, effects: ThemeEffect[]) => {
    const root = document.documentElement;
    for (const variable of themeVariableNames) root.style.removeProperty(variable);

    root.classList.toggle('theme-installed', !!theme);
    root.classList.toggle('theme-retro-windows', theme?.id === 'retro-windows');
    root.classList.toggle('theme-cursor', !!theme && effects.includes('retro-cursor'));

    if (theme) {
        for (const [name, value] of Object.entries(
            cssVariablesFor(theme.variants[mode], theme.variants.light, theme.variants.dark)
        )) {
            root.style.setProperty(name, value);
        }
    }
};

export const paletteForTheme = (theme: ThemePackage | undefined, mode: ThemeMode): Palette => {
    const basePalette = mode === 'dark' ? darkPalette : lightPalette;
    if (!theme) return basePalette;

    const colors = theme.variants[mode];
    const [background, surface, raised, hover, border] = colors.neutral;
    return createPalette({
        mode,
        primary: {
            main: colors.primary,
            dark: readableVariant(colors.primaryVariant, colors.onPrimary),
            contrastText: colors.onPrimary,
        },
        secondary: { main: colors.secondary, dark: colors.secondaryVariant, contrastText: colors.onSecondary },
        color: { primary: colors.text, links: colors.link, error: basePalette.color.error },
        neutral: {
            primary: background,
            secondary: surface,
            tertiary: raised,
            quaternary: hover,
            quinary: border,
        },
        background: { default: background, paper: surface },
        low: basePalette.low,
        moderate: basePalette.moderate,
        high: basePalette.high,
        critical: basePalette.critical,
    });
};
