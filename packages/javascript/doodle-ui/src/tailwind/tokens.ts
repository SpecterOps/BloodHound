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
import { common, dark, light, palette, text } from './colors';

const secondaryVariant2 = '#99a3ff';
const darkDataTableRowSelectedOutline = '#4A42B5';
const focusRingWidth = '2px';
const focusRingOffsetWidth = '1px';

/**
 * A single design token, mirroring a color variable in Figma.
 *
 * Each token is defined once here and generates:
 * - a CSS custom property per theme (see `createCssVariables`, used by the plugin)
 * - a Tailwind color referencing that variable (see `createTailwindColors`, used by the preset)
 *
 * Names are derived from the token's path in the tree, joined with '-', so the
 * Figma variable `status/error/main` lives at `tokens.status.error.main` and
 * produces `--status-error-main` and utilities like `bg-status-error-main`.
 */
export type Token = {
    /** Value in the light theme, emitted on `:root`. */
    light: string;
    /** Value in the dark theme, emitted on `.dark`. Omit when the token does not change in dark mode. */
    dark?: string;
    /** Emit only the CSS variable; do not generate a Tailwind color (and thus no utility classes). */
    cssVariableOnly?: boolean;
    /** Override the derived CSS variable name (without the `--` prefix). */
    varName?: string;
};

export type TokenGroup = {
    [key: string]: Token | TokenGroup;
};

const isToken = (node: Token | TokenGroup): node is Token => typeof (node as Token).light === 'string';

export const tokens = {
    // SHARED (same in light and dark)
    common: {
        dark: { light: common.dark },
        white: { light: common.white },
    },

    // MAIN
    primary: {
        /* DEFAULT and `main` are failsafes for classes using the optional '-main' suffix */
        DEFAULT: { light: light.primary.main, dark: dark.primary.main },
        main: { light: light.primary.main, dark: dark.primary.main },
        variant: { light: light.primary.variant, dark: dark.primary.variant },
    },
    secondary: {
        DEFAULT: { light: light.secondary.main, dark: dark.secondary.main },
        main: { light: light.secondary.main, dark: dark.secondary.main },
        variant: { light: light.secondary.variant, dark: dark.secondary.variant },
        'variant-2': { light: secondaryVariant2 },
    },
    tertiary: {
        // DEFAULT and `variant` hold legacy values; the new tokens are commented until adopted
        DEFAULT: { light: '#02c577' },
        // DEFAULT: { light: light.tertiary.main, dark: dark.tertiary.main },
        main: { light: light.tertiary.main, dark: dark.tertiary.main },
        variant: { light: '#5cc791' },
        // variant: { light: light.tertiary.variant, dark: dark.tertiary.variant },
    },
    disabled: { light: light.disabled, dark: dark.disabled },

    // NEUTRALS
    neutral: {
        50: { light: palette.neutral.light[50], dark: palette.neutral.dark[50] },
        // 100: { light: palette.neutral.light[100], dark: palette.neutral.dark[100] },
        200: { light: palette.neutral.light[200], cssVariableOnly: true },
        300: { light: palette.neutral.light[300], cssVariableOnly: true },
        400: { light: palette.neutral.light[400] },
        // 500 through 900 follow the same pattern and are pending adoption
    },

    // TEXT
    text: {
        main: { light: common.dark, dark: common.white },
        light: { light: text.light, dark: text.dark },
        // contrast: { light: common.white, dark: common.dark },
        disabled: { light: light.text.disabled, dark: common.disabled },
        // primary: { light: light.primary.main, dark: dark.primary.main },
        // secondary: { light: light.secondary.main, dark: dark.secondary.main },
    },

    // LINKS
    link: {
        DEFAULT: { light: light.secondary.main, dark: dark.secondary.main, varName: 'link-main' },
        hover: { light: light.secondary.variant, dark: dark.secondary.variant },
        // legacy `--link` variable, superseded by `--link-main`
        legacy: { light: '#1a30ff', dark: secondaryVariant2, cssVariableOnly: true, varName: 'link' },
    },

    // FOCUS (variables only; consumed by the `.focus-ring` utilities in the plugin)
    focus: {
        ring: {
            DEFAULT: { light: light.secondary.main, dark: dark.secondary.main, cssVariableOnly: true },
            width: { light: focusRingWidth, dark: focusRingWidth, cssVariableOnly: true },
            offset: {
                DEFAULT: { light: common.white, dark: palette.neutral.dark[50], cssVariableOnly: true },
                width: { light: focusRingOffsetWidth, dark: focusRingOffsetWidth, cssVariableOnly: true },
            },
        },
    },

    // STATUS
    status: {
        error: {
            main: { light: light.status.error.main, dark: dark.status.error.main },
            text: { light: light.status.error.text, dark: dark.status.error.text },
            fill: { light: light.status.error.fill, dark: dark.status.error.fill },
        },
        warning: {
            main: { light: light.status.warning.main, dark: dark.status.warning.main },
            text: { light: light.status.warning.text, dark: dark.status.warning.text },
            fill: { light: light.status.warning.fill, dark: dark.status.warning.fill },
        },
        success: {
            main: { light: light.status.success.main, dark: dark.status.success.main },
            text: { light: light.status.success.text, dark: dark.status.success.text },
            fill: { light: light.status.success.fill, dark: dark.status.success.fill },
        },
        info: {
            main: { light: light.status.info.main, dark: dark.status.info.main },
            text: { light: light.status.info.text, dark: dark.status.info.text },
            fill: { light: light.status.info.fill, dark: dark.status.info.fill },
        },
        indeterminate: {
            fill: { light: light.status.indeterminate.fill, dark: dark.status.indeterminate.fill },
        },
    },

    // Components/Button
    'secondary-btn': {
        fill: { light: palette.neutral.light[300], dark: palette.neutral.dark[700] },
        'active-fill': { light: palette.neutral.light[400], dark: palette.neutral.dark[700] },
    },
    'tertiary-btn': {
        border: { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
    },
    'transparent-btn': {
        border: { light: palette.neutral.light[400], dark: palette.neutral.dark[700] },
    },
    'icon-btn': {
        fill: { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
    },
    'btn-disabled': {
        fill: { light: palette.neutral.light[200], dark: palette.neutral.dark[700] },
    },
    'toggle-btn': {
        fill: { light: common.white, dark: common.dark },
        border: { light: palette.neutral.light[500], dark: palette.neutral.dark[600] },
    },
    'toggle-group': {
        fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
    },
    checkbox: {
        border: { light: common.black, dark: common.white },
        'unchecked-fill': { light: common.white, dark: common.dark },
        fill: { light: common.dark, dark: common.white },
        check: { light: common.white, dark: common.dark },
    },

    // Components/Input
    input: {
        fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
        border: {
            default: { light: palette.neutral.dark[700], dark: palette.neutral.light[400] },
            focus: { light: light.secondary.main, dark: secondaryVariant2 },
        },
    },

    // Components/Textarea
    textarea: {
        fill: { light: common.white, dark: palette.neutral.dark[700] },
        border: {
            default: { light: palette.neutral.light[400], dark: palette.neutral.dark[700] },
            hover: { light: light.secondary.main, dark: secondaryVariant2 },
        },
    },

    // Components/RadioGroup
    radio: {
        'label-focus-fill': { light: palette.neutral.light[200], dark: palette.neutral.dark[600] },
        border: {
            default: { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
            hover: { light: light.secondary.main, dark: secondaryVariant2 },
        },
        disabled: {
            border: { light: palette.neutral.light[300], dark: palette.neutral.dark[600] },
            fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
        },
        indicator: {
            fill: { light: light.primary.main, dark: secondaryVariant2 },
        },
    },

    // Components/Input/Selectors
    select: {
        trigger: {
            fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
            'placeholder-text': { light: palette.neutral.light[400], dark: palette.neutral.dark[700] },
            'outlined-fill': { light: common.white, dark: palette.neutral.dark[50] },
        },
        border: {
            default: { light: palette.neutral.dark[700], dark: palette.neutral.light[400] },
            focus: { light: light.secondary.main, dark: secondaryVariant2 },
        },
        content: {
            border: { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
            fill: { light: common.white, dark: palette.neutral.dark[400] },
        },
        'item-checked-text': { light: light.primary.main, dark: secondaryVariant2 },
        separator: {
            fill: { light: palette.neutral.light[200], dark: palette.neutral.light[200] },
        },
    },
    dropdown: {
        trigger: {
            border: { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
        },
        popover: {
            border: { light: palette.neutral.light[400], dark: palette.neutral.dark[700] },
            fill: { light: common.white, dark: common.dark },
        },
        option: {
            'hover-fill': { light: palette.neutral.light[300], dark: palette.neutral.dark[600] },
            'disabled-fill': { light: palette.neutral.light[300], dark: palette.neutral.dark[600] },
        },
        tooltip: {
            fill: { light: palette.neutral.light[300], dark: palette.neutral.dark[600] },
        },
    },
    switch: {
        fill: { light: palette.neutral.dark[700], dark: common.white },
        'disabled-fill': { light: palette.neutral.light[300], dark: common.disabled },
        thumb: {
            fill: { light: palette.neutral.light[50], dark: palette.neutral.dark[50] },
            'disabled-fill': { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
        },
    },

    // Components/Data Display/DataTable
    'data-table': {
        fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
        header: {
            fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
        },
        row: {
            'even-fill': { light: palette.neutral.light[200], dark: palette.neutral.dark[500] },
            'odd-fill': { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
            'hover-fill': { light: palette.neutral.light[300], dark: palette.neutral.dark[600] },
            'selected-outline': { light: light.primary.main, dark: darkDataTableRowSelectedOutline },
        },
    },

    // Components/Data Display/Badge & Chip
    badge: {
        primary: {
            fill: { light: light.badge.primary.fill, dark: dark.badge.primary.fill },
            outline: { light: light.badge.primary.outline, dark: dark.badge.primary.outline },
        },
        secondary: {
            fill: { light: light.badge.secondary.fill, dark: dark.badge.secondary.fill },
            outline: { light: light.badge.secondary.outline, dark: dark.badge.secondary.outline },
        },
        grey: {
            fill: { light: light.badge.grey.fill, dark: dark.badge.grey.fill },
            outline: { light: light.badge.grey.outline, dark: dark.badge.grey.outline },
        },
        red: {
            fill: { light: light.badge.red.fill, dark: dark.badge.red.fill },
            outline: { light: light.badge.red.outline, dark: dark.badge.red.outline },
        },
        orange: {
            fill: { light: light.badge.orange.fill, dark: dark.badge.orange.fill },
            outline: { light: light.badge.orange.outline, dark: dark.badge.orange.outline },
        },
        green: {
            fill: { light: light.badge.green.fill, dark: dark.badge.green.fill },
            outline: { light: light.badge.green.outline, dark: dark.badge.green.outline },
        },
        blue: {
            fill: { light: light.badge.blue.fill, dark: dark.badge.blue.fill },
            outline: { light: light.badge.blue.outline, dark: dark.badge.blue.outline },
        },
    },

    // Figma variables not yet adopted. Uncommenting a token emits its CSS
    // variable and utility classes.
    //
    // risk: {
    //     critical: { light: palette.purple.A300 },
    //     high: { light: palette.red.A300 },
    //     moderate: { light: palette.brown[300] },
    //     low: { light: palette.yellow.A300 },
    //     mitigated: { light: palette.green.A300 },
    //     resolved: { light: palette['light-blue'].A300 },
    //     accepted: { light: palette.neutral.light[300] },
    //     text: { light: common.black },
    // },
    // elevation: {
    //     0: { light: elevation.light[0], dark: elevation.dark[0] },
    //     1: { light: elevation.light[1], dark: elevation.dark[1] },
    //     2: { light: elevation.light[2], dark: elevation.dark[2] },
    //     3: { light: elevation.light[3], dark: elevation.dark[3] },
    //     4: { light: elevation.light[4], dark: elevation.dark[4] },
    //     5: { light: elevation.light[5], dark: elevation.dark[5] },
    // },
    // 'bhe-main': { light: light.primary.main, dark: dark.primary.main },
    // 'sp-main': { light: light.primary.main, dark: dark.primary.main },
    // 'bhce-main': { light: light['bhce-main'], dark: dark['bhce-main'] },
    // 'logo-neutral': { light: common.black, dark: common.white },
    // brand: {
    //     primary: {
    //         'dark-purple': { light: light.primary.variant, dark: dark.primary.variant },
    //         'deep-purple': { light: light.primary.main, dark: dark.primary.main },
    //         'medium-blue': { light: light.secondary.main, dark: dark.secondary.main },
    //         'highlight-neon-blue': { light: common['neon-blue'] },
    //         'highlight-orange': { light: common.orange },
    //     },
    //     secondary: {
    //         'medium-purple': { light: common['medium-purple'] },
    //         'light-purple': { light: common['light-purple'] },
    //         'light-blue-gray': { light: common['light-blue-gray'] },
    //         'highlight-light-blue': { light: light.secondary.main },
    //         'highlight-green': { light: light.tertiary.main, dark: dark.tertiary.main },
    //     },
    // },
    // input: {
    //     label: { light: common.dark, dark: common.white },
    //     'fill-disabled': { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
    //     border: {
    //         hover: { light: light.secondary.main, dark: dark.secondary.main },
    //         disabled: { light: palette.neutral.light[900], dark: palette.neutral.dark[900] },
    //     },
    //     'placeholder-text': { light: text.placeholder, dark: dark.input.placeholder },
    // },
    // 'selector-disable-fill': { light: common.white },
    // chip: {
    //     indeterminate: {
    //         DEFAULT: { light: palette.neutral.light[300], dark: palette.neutral.dark[500] },
    //         hover: { light: palette.neutral.light[600], dark: palette.neutral.dark[300] },
    //     },
    //     outline: {
    //         fill: { light: common.white, dark: elevation.dark[1] },
    //         hover: { light: palette.neutral.light[100], dark: palette.neutral.dark[200] },
    //         active: { light: palette.neutral.light[300], dark: palette.neutral.dark[400] },
    //     },
    // },
    // 'menu-bg': { light: elevation.light[0], dark: elevation.dark[2] },
    // sidenav: {
    //     bg: {
    //         DEFAULT: { light: palette.neutral.light[50], dark: palette.neutral.dark[300] },
    //         hover: { light: palette.neutral.light[200], dark: palette.neutral.dark[200] },
    //         active: { light: palette.neutral.light[300], dark: palette.neutral.dark[600] },
    //     },
    // },
    // 'nav-control-btn': {
    //     DEFAULT: { light: palette.neutral.light[300], dark: palette.neutral.dark[300] },
    //     hover: { light: palette.neutral.light[600], dark: palette.neutral.dark[500] },
    //     focus: { light: palette.neutral.light[500], dark: palette.neutral.dark[400] },
    // },
    // 'nav-item': {
    //     default: { light: common.dark, dark: common.white },
    //     hover: { light: light.secondary.main, dark: dark.secondary.main },
    //     active: { light: light.primary.main, dark: dark.primary.main },
    // },
    // 'platform-badge-bg': { light: elevation.light[2], dark: elevation.dark[2] },
    // edge: {
    //     color: { light: light.edge.color, dark: dark.edge.color },
    //     'label-bg': { light: palette.neutral.light[200], dark: palette.neutral.dark[500] },
    //     'label-text': { light: common.dark, dark: common.white },
    // },
    // icon: {
    //     DEFAULT: { light: common.dark, dark: common.white },
    //     contrast: { light: common.white, dark: common.dark },
    //     disabled: { light: palette.grey[700], dark: common.disabled },
    // },
    // divider: { light: palette.neutral.light[500], dark: palette.neutral.dark[500] },
} satisfies TokenGroup;

/**
 * Legacy tokens kept for backwards compatibility. Defined separately so it is
 * obvious what can be deleted once usages have migrated to `tokens`.
 */
export const legacyTokens = {
    // same as palette.neutral.dark[50] (light) / common.white (dark)
    contrast: { light: '#121212', dark: '#ffffff' },
    error: { light: '#b44641', dark: '#e9827c' },

    'neutral-1': { light: '#ffffff', dark: '#121212' },
    'neutral-2': { light: '#f4f4f4', dark: '#222222' },
    'neutral-3': { light: '#e3e7ea', dark: '#272727' },
    'neutral-4': { light: '#dadee1', dark: '#2c2c2c' },
    'neutral-5': { light: '#cacfd3', dark: '#2e2e2e' },

    'neutral-light-1': { light: '#ffffff' },
    'neutral-light-2': { light: '#f4f4f4' },
    'neutral-light-3': { light: '#e3e7ea' },
    'neutral-light-4': { light: '#dadee1' },
    'neutral-light-5': { light: '#cacfd3' },

    'neutral-dark-1': { light: '#121212' },
    'neutral-dark-2': { light: '#222222' },
    'neutral-dark-3': { light: '#272727' },
    'neutral-dark-4': { light: '#2c2c2c' },
    'neutral-dark-5': { light: '#2e2e2e' },
} satisfies TokenGroup;

/**
 * Flattens a token tree into the CSS custom properties for one theme, e.g.
 * `tokens.status.error.main` -> `'--status-error-main': '#B44641'`.
 * A `DEFAULT` key maps to the name of its parent group. Tokens without a
 * `dark` value are omitted from the dark theme and inherit from `:root`.
 */
export const createCssVariables = (
    group: TokenGroup,
    theme: 'light' | 'dark',
    path: string[] = []
): Record<string, string> => {
    const variables: Record<string, string> = {};
    for (const [key, node] of Object.entries(group)) {
        const segments = key === 'DEFAULT' ? path : [...path, key];
        if (isToken(node)) {
            const value = node[theme];
            if (value !== undefined) {
                variables[`--${node.varName ?? segments.join('-')}`] = value;
            }
        } else {
            Object.assign(variables, createCssVariables(node, theme, segments));
        }
    }
    return variables;
};

type TailwindColors = {
    [key: string]: string | TailwindColors;
};

/**
 * Maps a token tree into a Tailwind `theme.extend.colors` object in which every
 * leaf references its CSS variable. Nested groups produce the same class names
 * as flat kebab-case entries, so `tokens.status.error.main` yields utilities
 * like `bg-status-error-main` backed by `var(--status-error-main)`.
 */
export const createTailwindColors = (group: TokenGroup, path: string[] = []): TailwindColors => {
    const colors: TailwindColors = {};
    for (const [key, node] of Object.entries(group)) {
        const segments = key === 'DEFAULT' ? path : [...path, key];
        if (isToken(node)) {
            if (!node.cssVariableOnly) {
                colors[key] = `var(--${node.varName ?? segments.join('-')})`;
            }
        } else {
            const nested = createTailwindColors(node, segments);
            if (Object.keys(nested).length > 0) {
                colors[key] = nested;
            }
        }
    }
    return colors;
};
