// Copyright 2024 Specter Ops, Inc.
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

import type { Palette } from '@mui/material/styles/createPalette';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { ThemePresetConfig } from '../constants';

/**
 * This function sets the name of our current theme as a class on the html document root. This will ensure the correct styles are applied to components attached elsewhere in the DOM, such as modals and popover menus.
 *
 * @param value - can be 'dark' or 'light'
 *
 * @returns the name of the currently set class as a string
 */
export const setRootClass = (value: 'dark' | 'light') => {
    const root = window.document.documentElement;
    root.classList.remove('dark', 'light');
    root.classList.add(value);
    return value;
};

const neutralSteps = ['primary', 'secondary', 'tertiary', 'quaternary', 'quinary'] as const;

// Maps a MUI palette onto the doodle-ui color variables for the active mode.
const getModeColorVariables = (palette: Palette): Record<string, string> => {
    const variables: Record<string, string> = {
        '--primary': palette.primary.main,
        '--primary-main': palette.primary.main,
        '--primary-variant': palette.primary.dark,
        '--bhe-main': palette.primary.main,
        '--radio-indicator-fill': palette.primary.main,
        '--select-item-checked-text': palette.primary.main,
        '--data-table-row-selected-outline': palette.primary.main,
        '--secondary': palette.secondary.main,
        '--secondary-main': palette.secondary.main,
        '--secondary-variant': palette.secondary.dark,
        '--focus-ring': palette.secondary.main,
        '--checkbox-hover': palette.secondary.main,
        '--input-outlined-border-hover': palette.secondary.main,
        '--textarea-border-hover': palette.secondary.main,
        '--radio-border-hover': palette.secondary.main,
        '--select-border-focus': palette.secondary.main,
        '--text-main': palette.color.primary,
        '--link': palette.color.links,
        '--link-main': palette.color.links,
        '--error': palette.color.error,
    };

    neutralSteps.forEach((step, index) => {
        variables[`--neutral-${index + 1}`] = palette.neutral[step];
    });

    return variables;
};

/**
 * Builds the CSS variables needed for Tailwind/doodle-ui styled components to follow a theme preset.
 * Fonts are always included. Colors are only included when the preset overrides the doodle-ui design tokens.
 *
 * @param preset - the selected theme preset
 * @param darkMode - whether dark mode is enabled
 *
 * @returns a map of CSS variable names to values
 */
export const getThemePresetCssVariables = (preset: ThemePresetConfig, darkMode: boolean): Record<string, string> => {
    const variables: Record<string, string> = {
        '--font-body': preset.bodyFontFamily,
        '--font-heading': preset.headingFontFamily,
    };

    if (!preset.overridesDoodleColors) return variables;

    // Components often pair fixed light/dark tokens (e.g. `bg-neutral-light-2 dark:bg-neutral-dark-2`), so both
    // palettes are mapped regardless of the active mode.
    neutralSteps.forEach((step, index) => {
        variables[`--neutral-light-${index + 1}`] = preset.lightPalette.neutral[step];
        variables[`--neutral-dark-${index + 1}`] = preset.darkPalette.neutral[step];
    });

    return { ...variables, ...getModeColorVariables(darkMode ? preset.darkPalette : preset.lightPalette) };
};

let appliedThemeVariableNames: string[] = [];

/**
 * Applies a theme preset's CSS variables as inline styles on the html document root, so they take precedence over
 * the doodle-ui `:root` and `.dark` defaults. Variables set by a previous preset that the new one does not set are
 * removed, restoring the doodle-ui defaults.
 *
 * @param preset - the selected theme preset
 * @param darkMode - whether dark mode is enabled
 */
export const applyThemePresetCssVariables = (preset: ThemePresetConfig, darkMode: boolean) => {
    const rootStyle = window.document.documentElement.style;
    const variables = getThemePresetCssVariables(preset, darkMode);

    appliedThemeVariableNames.filter((name) => !(name in variables)).forEach((name) => rootStyle.removeProperty(name));

    Object.entries(variables).forEach(([name, value]) => rootStyle.setProperty(name, value));

    appliedThemeVariableNames = Object.keys(variables);
};

/**
 * Utility function for conditionally constructing className strings and merging the result.
 *
 * @param inputs - any number of valid clsx statements. For reference: [clsx docs](https://github.com/lukeed/clsx#readme)
 *
 * @returns a merged class list as a string. For more information about how merging tailwind classes works: [twMerge docs](https://github.com/dcastil/tailwind-merge/blob/v2.5.4/docs/what-is-it-for.md)
 */
export const cn = (...inputs: ClassValue[]) => {
    return twMerge(clsx(inputs));
};
