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
import { brand, common, dark, light, palette } from './colors';

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
    // TODO stays here
    common: {
        dark: { light: palette.neutral.dark[50], dark: common.white },
        main: { light: palette.black, dark: palette.white },
        disabled: { light: common.disabled.light, dark: common.disabled.dark },
        contrast: { light: common.white, dark: common.dark },
        white: { light: common.white }, // TODO does this need to exist?
        divider: { light: palette.neutral.light[500], dark: palette.neutral.dark[500] },
    },
    // MAIN
    // TODO stays here
    primary: {
        /* DEFAULT and `main` are failsafes for classes using the optional '-main' suffix */
        DEFAULT: { light: brand.purple.dark, dark: brand.purple.light },
        main: { light: brand.purple.dark, dark: brand.purple.light },
        // TODO can dark variant be the same as purple variant '#99a3ff'?
        variant: { light: '#0D0A30', dark: '#8D8BF8' },
    },
    // TODO stays here
    secondary: {
        DEFAULT: { light: brand.purple.medium, dark: brand.blue.medium },
        main: { light: brand.purple.medium, dark: brand.blue.medium },
        variant: { light: '#3729BB', dark: '#4E95FF' },
        'variant-2': { light: brand.purple.variant },
    },
    // TODO should tertiary even be here?
    tertiary: {
        // DEFAULT and `variant` hold legacy values; the new tokens are commented until adopted
        DEFAULT: { light: '#02c577' },
        // DEFAULT: { light: light.tertiary.main, dark: dark.tertiary.main },
        main: { light: light.tertiary.main, dark: dark.tertiary.main },
        variant: { light: '#5cc791' },
        // variant: { light: light.tertiary.variant, dark: dark.tertiary.variant },
    },
    // TODO doesn't need to exist - can be removed
    disabled: { light: common.disabled.light, dark: common.disabled.dark }, // -> use common.disabled

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
    // TODO stays here
    text: {
        main: { light: common.dark, dark: common.white }, // -> use common.dark
        light: { light: '#505050', dark: '#CDCDCD' },
        // contrast: { light: common.white, dark: common.dark }, // -> use common.contrast
        disabled: { light: common.disabled.light, dark: common.disabled.dark }, // -> use common.disabled
        // primary: { light: light.primary.main, dark: dark.primary.main }, // -> use primary.main
        // secondary: { light: light.secondary.main, dark: dark.secondary.main }, // -> use secondary.main
    },

    // LINKS
    // moved to Link component - can be removed
    link: {
        DEFAULT: { light: brand.purple.medium, dark: brand.blue.medium, varName: 'link-main' },
        hover: { light: '#3729BB', dark: '#4E95FF' },
        // legacy `--link` variable, superseded by `--link-main`
        legacy: { light: '#1a30ff', dark: brand.purple.variant, cssVariableOnly: true, varName: 'link' },
    },

    // FOCUS (variables only; consumed by the `.focus-ring` utilities in the plugin)
    // TODO stays here
    focus: {
        ring: {
            DEFAULT: { light: brand.purple.medium, dark: brand.blue.medium, cssVariableOnly: true },
            width: { light: focusRingWidth, dark: focusRingWidth, cssVariableOnly: true },
            offset: {
                DEFAULT: { light: common.white, dark: palette.neutral.dark[50], cssVariableOnly: true },
                width: { light: focusRingOffsetWidth, dark: focusRingOffsetWidth, cssVariableOnly: true },
            },
        },
    },

    // STATUS
    // TODO stays here
    status: {
        error: {
            // same as theme.red.medium
            main: { light: '#B44641', dark: palette.red[200] },
            text: { light: '#5F2120', dark: palette.red[100] },
            fill: { light: palette.red[50], dark: '#5F2120' },
        },
        warning: {
            main: { light: palette.orange[900], dark: palette.orange[500] },
            text: { light: '#6D3900', dark: palette.orange[100] },
            fill: { light: '#FFE1D1', dark: '#452F16' },
        },
        success: {
            main: { light: palette.green[600], dark: palette.green[200] },
            text: { light: '#1E4620', dark: palette.green[100] },
            fill: { light: palette.green[50], dark: '#1E4620' },
        },
        info: {
            main: { light: palette['light-blue'][700], dark: palette['light-blue'][200] },
            text: { light: '#004465', dark: palette['light-blue'][100] },
            fill: { light: palette['light-blue'][50], dark: '#103440' },
        },
        indeterminate: {
            fill: { light: palette.neutral.dark[400], dark: palette.white },
        },
    },

    // Components/Button
    // moved to Button component - can be removed
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

    // moved to checkbox component - can be removed
    checkbox: {
        border: { light: common.black, dark: common.white },
        'unchecked-fill': { light: common.white, dark: common.dark },
        fill: { light: common.dark, dark: common.white },
        check: { light: common.white, dark: common.dark },
    },

    // Components/Input
    // moved to input component - can be removed
    input: {
        fill: {
            light: palette.neutral.light[100],
            dark: palette.neutral.dark[400],
        },
        border: {
            default: { light: palette.neutral.dark[700], dark: palette.neutral.light[400] },
            focus: { light: brand.purple.medium, dark: brand.purple.variant },
        },
        //     label: { light: common.dark, dark: common.white },
        //     'fill-disabled': { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
        //     border: {
        //         hover: { light: light.secondary.main, dark: dark.secondary.main },
        //         disabled: { light: palette.neutral.light[900], dark: palette.neutral.dark[900] },
        //     },
        //     'placeholder-text': { light: text.placeholder, dark: dark.input.placeholder },
    },

    // Components/Textarea
    // moved to textarea component - can be removed
    textarea: {
        fill: { light: common.white, dark: palette.neutral.dark[700] },
        border: {
            default: { light: palette.neutral.light[400], dark: palette.neutral.dark[700] },
            hover: { light: brand.purple.medium, dark: brand.purple.variant },
        },
    },

    // Components/RadioGroup
    // moved to radioGroup component - can be removed
    radio: {
        'label-focus-fill': { light: palette.neutral.light[200], dark: palette.neutral.dark[600] },
        border: {
            default: { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
            hover: { light: brand.purple.medium, dark: brand.purple.variant },
        },
        disabled: {
            border: { light: palette.neutral.light[300], dark: palette.neutral.dark[600] },
            fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
        },
        indicator: {
            fill: { light: brand.purple.dark, dark: brand.purple.variant },
        },
    },

    // Components/Input/Selectors
    // moved to select component - can be removed
    select: {
        trigger: {
            fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
            'placeholder-text': { light: palette.neutral.light[400], dark: palette.neutral.dark[700] },
            'outlined-fill': { light: common.white, dark: palette.neutral.dark[50] },
        },
        border: {
            default: { light: palette.neutral.dark[700], dark: palette.neutral.light[400] },
            focus: { light: brand.purple.medium, dark: brand.purple.variant },
        },
        content: {
            border: { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
            fill: { light: common.white, dark: palette.neutral.dark[400] },
        },
        'item-checked-text': { light: brand.purple.dark, dark: brand.purple.variant },
        separator: {
            fill: { light: palette.neutral.light[200], dark: palette.neutral.light[200] },
        },
    },

    // TODO move to dropdownSelector component *** located in BH-Shared-UI
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

    // moved to switch component - can be removed
    switch: {
        fill: { light: palette.neutral.dark[700], dark: palette.white },
        'disabled-fill': { light: palette.neutral.light[300], dark: common.disabled.dark },
        thumb: {
            fill: { light: palette.neutral.light[50], dark: palette.neutral.dark[50] },
            'disabled-fill': { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
        },
    },

    // Components/Data Display/DataTable
    // moved to dataTable component - can be removed
    'data-table': {
        fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
        header: {
            fill: { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
        },
        row: {
            'even-fill': { light: palette.neutral.light[200], dark: palette.neutral.dark[500] },
            'odd-fill': { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
            'hover-fill': { light: palette.neutral.light[300], dark: palette.neutral.dark[600] },
            'selected-outline': { light: brand.purple.dark, dark: '#4A42B5' },
        },
    },

    // Figma variables not yet adopted. Uncommenting a token emits its CSS
    // variable and utility classes.
    //
    // // moved to riskBadge component - can be removed
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
    // TODO - stays here
    // icon: {
    //     DEFAULT: { light: common.dark, dark: common.white },
    //     contrast: { light: common.white, dark: common.dark },
    //     disabled: { light: palette.grey[700], dark: common.disabled },
    // },
} satisfies TokenGroup;

/**
 * Legacy tokens kept for backwards compatibility. Defined separately so it is
 * obvious what can be deleted once usages have migrated to `tokens`.
 */
export const legacyTokens = {
    // same as palette.neutral.dark[50] (light) / common.white (dark)
    contrast: { light: '#121212', dark: '#ffffff' },
    error: { light: '#b44641', dark: '#e9827c' },
    // same as status.error.medium and theme.red.medium

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
            // TODO address light and dark mode first
            // remove mode from
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
