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

import { Theme, ThemeOptions } from '@mui/material/styles';
import createPalette, { Palette } from '@mui/material/styles/createPalette';
import { makeStyles } from '@mui/styles';
import { ActiveDirectoryKindProperties, AzureKindProperties, CommonKindProperties } from './graphSchema';
import { BaseExploreLayoutOptions, MappedStringLiteral } from './types';
import { addOpacityToHex } from './utils/colors';

// Max and min length requirements for creating/updating a user
export const MAX_NAME_LENGTH = 319;
export const MIN_NAME_LENGTH = 2;
export const MAX_DESCRIPTION_LENGTH = 500;

export const MAX_EMAIL_LENGTH = 319;

export const NODE_GRAPH_RENDER_LIMIT = 1000;

export const ZERO_VALUE_API_DATE = '0001-01-01T00:00:00Z';

// These tags are values associated with the `system_tags` property of a node
export const OWNED_OBJECT_TAG = 'owned';
export const TIER_ZERO_TAG = 'admin_tier_0';
// These tags are values associated with the new Tiering Management kind approach
export const TAG_TIER_ZERO_AGT = 'Tag_Tier_Zero';
export const TAG_OWNED_AGT = 'Tag_Owned';

// These labels are used as display values
export const TIER_ZERO_LABEL = 'Admin Tier Zero';
export const HIGH_VALUE_LABEL = 'High Value';

// Snackbar duration values
export const SNACKBAR_DURATION = 5000;
export const SNACKBAR_DURATION_LONG = 15000;

export const useStyles = makeStyles(() => ({
    applicationContainer: {
        display: 'flex',
        position: 'relative',
        height: '100%',
        overflow: 'hidden',
    },
}));

const focusRingStyles = (palette: Palette) => ({
    outline: `2px solid ${palette.color.links}`,
    outlineOffset: '2px',
});

const inheritFocusedIconStyles = {
    '& svg': {
        color: 'inherit',
        fill: 'currentColor',
    },
    '& svg *': {
        color: 'inherit',
        fill: 'currentColor',
    },
};

export const themedComponents = (palette: Palette): ThemeOptions['components'] => ({
    MuiButtonBase: {
        styleOverrides: {
            root: {
                '&.Mui-focusVisible': {
                    ...focusRingStyles(palette),
                    ...inheritFocusedIconStyles,
                },
            },
        },
    },
    MuiAccordionSummary: {
        styleOverrides: {
            root: {
                flexDirection: 'row-reverse',
                '&.Mui-focusVisible': focusRingStyles(palette),
            },
            content: {
                marginRight: '4px',
            },
        },
    },
    MuiLink: {
        styleOverrides: {
            root: {
                color: palette.color.links,
                borderRadius: '2px',
                '&:focus-visible': {
                    ...focusRingStyles(palette),
                    textDecoration: 'underline',
                    textDecorationThickness: '2px',
                    textUnderlineOffset: '2px',
                },
            },
        },
    },
    MuiInputLabel: {
        styleOverrides: {
            root: {
                '&.Mui-focused': {
                    color: palette.color.links,
                },
            },
        },
    },
    MuiTextField: {
        styleOverrides: {
            root: {
                '&:hover .MuiInputBase-root .MuiOutlinedInput-notchedOutline': {
                    borderColor: palette.color.links,
                },
                '& .MuiInputBase-root.Mui-focused .MuiOutlinedInput-notchedOutline': {
                    borderColor: palette.color.links,
                },
            },
        },
    },
    MuiInput: {
        styleOverrides: {
            underline: {
                '&:after': {
                    borderBottom: `2px solid ${palette.color.links}`,
                },
                '&:hover:not($disabled):not($focused):not($error):before': {
                    borderBottom: `2px solid ${palette.color.links}`,
                },
            },
        },
    },
    MuiDialog: {
        defaultProps: {
            ...defaultPortalContainer,
        },
        styleOverrides: {
            root: {
                '& .MuiPaper-root': {
                    backgroundImage: 'unset',
                    backgroundColor: palette.neutral.secondary,
                },
            },
        },
    },
    MuiMenu: {
        defaultProps: {
            ...defaultPortalContainer,
        },
    },
    MuiAutocomplete: {
        defaultProps: {
            componentsProps: {
                popper: {
                    ...defaultPortalContainer,
                },
            },
        },
        styleOverrides: {
            option: {
                '&.Mui-focused, &[aria-selected="true"].Mui-focused': {
                    backgroundColor: addOpacityToHex(palette.color.links, 16),
                    boxShadow: `inset 3px 0 0 ${palette.color.links}`,
                },
            },
        },
    },
    MuiDialogActions: {
        styleOverrides: {
            root: {
                padding: '16px 24px',
            },
        },
    },
    MuiPopover: {
        defaultProps: {
            ...defaultPortalContainer,
        },
        styleOverrides: {
            root: {
                '& .MuiPaper-root': {
                    backgroundImage: 'unset',
                },
            },
        },
    },
    MuiCheckbox: {
        styleOverrides: {
            root: {
                '&.Mui-focusVisible': {
                    ...focusRingStyles(palette),
                    borderRadius: '4px',
                },
                '& svg': {
                    color: palette.color.primary,
                },
            },
        },
    },
    MuiRadio: {
        styleOverrides: {
            root: {
                '&.Mui-focusVisible': {
                    ...focusRingStyles(palette),
                    borderRadius: '50%',
                },
            },
        },
    },
    MuiSwitch: {
        styleOverrides: {
            root: {
                '&:has(.Mui-focusVisible)': {
                    ...focusRingStyles(palette),
                    borderRadius: '999px',
                },
            },
        },
    },
    MuiTabs: {
        styleOverrides: {
            root: {
                '& .MuiTab-labelIcon': {
                    color: palette.color.links,
                },
                '& .MuiButtonBase-root.Mui-selected': {
                    color: palette.color.links,
                },
                '& .MuiTab-labelIcon:not(.Mui-selected)': {
                    color: palette.color.primary,
                },
                '& .MuiTabs-indicator': {
                    backgroundColor: palette.color.links,
                },
                '& .Mui-selected > svg': {
                    color: palette.color.links,
                },
                '& :not(.Mui-selected) > svg': {
                    color: palette.color.primary,
                },
            },
        },
    },
    MuiTab: {
        styleOverrides: {
            root: {
                '&.Mui-focusVisible': {
                    ...focusRingStyles(palette),
                    ...inheritFocusedIconStyles,
                },
            },
        },
    },
    MuiMenuItem: {
        styleOverrides: {
            root: {
                '&.Mui-focusVisible, &:focus-visible': {
                    backgroundColor: addOpacityToHex(palette.color.links, 16),
                    boxShadow: `inset 3px 0 0 ${palette.color.links}`,
                    ...inheritFocusedIconStyles,
                },
            },
        },
    },
    MuiTableSortLabel: {
        styleOverrides: {
            root: {
                borderRadius: '2px',
                '&.Mui-focusVisible, &:focus-visible': focusRingStyles(palette),
            },
        },
    },
    MuiAlert: {
        styleOverrides: {
            root: {
                '&.MuiAlert-standardWarning': {
                    backgroundColor: addOpacityToHex(palette.warning.main, 20),
                },
                '&.MuiAlert-standardInfo': {
                    backgroundColor: addOpacityToHex(palette.info.main, 20),
                },
                '&.MuiAlert-standardError': {
                    backgroundColor: addOpacityToHex(palette.error.main, 20),
                },
            },
        },
    },
    MuiLinearProgress: {
        styleOverrides: {
            root: {
                backgroundColor: addOpacityToHex(palette.primary.main, 40),
                '& .MuiLinearProgress-barColorPrimary': {
                    backgroundColor: palette.primary.main,
                },
            },
        },
    },
    MuiTableContainer: {
        styleOverrides: {
            root: {
                backgroundImage: 'unset',
            },
        },
    },
    MuiPaper: {
        styleOverrides: {
            root: {
                backgroundImage: 'unset',
            },
        },
    },
});

export const lightPalette = createPalette({
    mode: 'light',
    primary: {
        main: '#33318F',
        dark: '#261F7A',
    },
    secondary: {
        main: '#1A30FF',
        dark: '#0524F0',
    },
    color: {
        primary: '#1D1B20',
        links: '#1A30FF',
        error: '#B44641',
    },
    neutral: {
        primary: '#FFFFFF',
        secondary: '#F4F4F4',
        tertiary: '#E3E7EA',
        quaternary: '#DADEE1',
        quinary: '#CACFD3',
    },
    background: {
        paper: '#fafafa',
        default: '#e4e9eb',
    },
    low: 'rgb(255, 195, 15)',
    moderate: 'rgb(255, 97, 66)',
    high: 'rgb(205, 0, 117)',
    critical: 'rgb(76, 29, 143)',
});

export const darkPalette = createPalette({
    mode: 'dark',
    primary: {
        main: '#33318F',
        dark: '#261F7A',
    },
    secondary: {
        main: '#1A30FF',
        dark: '#0524F0',
    },
    color: {
        primary: '#FFFFFF',
        links: '#99A3FF',
        error: '#E9827C',
    },
    neutral: {
        primary: '#121212',
        secondary: '#222222',
        tertiary: '#272727',
        quaternary: '#2C2C2C',
        quinary: '#2E2E2E',
    },
    background: {
        paper: '#211F26',
        default: '#121212',
    },
    low: 'rgb(255, 195, 15)',
    moderate: 'rgb(255, 97, 66)',
    high: 'rgb(205, 0, 117)',
    critical: 'rgb(76, 29, 143)',
});

// Font stacks mirror the doodle-ui Tailwind preset: Figtree for body text and Nunito Sans for headings.
const bodyFontFamily = 'Figtree, "Segoe UI", Helvetica, Arial, sans-serif';
const headingFontFamily = '"Nunito Sans", "Avenir Next", "Segoe UI", Helvetica, Arial, sans-serif';

// Builds a typography config from a body and heading font stack so theme presets can swap
// typefaces while keeping the shared sizing, weight, and spacing scale.
const buildTypography = (bodyFont: string, headingFont: string): Partial<Theme['typography']> => ({
    fontFamily: bodyFont,
    h1: {
        fontFamily: headingFont,
        fontWeight: 600,
        fontSize: '1.8rem',
        lineHeight: 2,
        letterSpacing: 0,
    },
    h2: {
        fontFamily: headingFont,
        fontWeight: 600,
        fontSize: '1.5rem',
        lineHeight: 1.5,
        letterSpacing: 0,
    },
    h3: {
        fontFamily: headingFont,
        fontWeight: 600,
        fontSize: '1.2rem',
        lineHeight: 1.25,
        letterSpacing: 0,
    },
    h4: {
        fontFamily: headingFont,
        fontWeight: 600,
        fontSize: '1.25rem',
        lineHeight: 1.5,
        letterSpacing: 0,
    },
    h5: {
        fontFamily: headingFont,
        fontWeight: 700,
        fontSize: '1.125rem',
        lineHeight: 1.5,
        letterSpacing: 0.25,
    },
    h6: {
        fontFamily: headingFont,
        fontWeight: 700,
        fontSize: '1.0rem',
        lineHeight: 1.5,
        letterSpacing: 0.25,
    },
});

export const typography: Partial<Theme['typography']> = buildTypography(bodyFontFamily, headingFontFamily);

// Font stacks for the alternate theme presets. Comic Sans and the serif stack both rely on
// fonts that ship with the operating system, so no additional @fontsource imports are needed.
const comicSansFontFamily = '"Comic Sans MS", "Comic Sans", "Chalkboard SE", "Comic Neue", cursive';
const serifFontFamily = 'Georgia, "Times New Roman", "Noto Serif", serif';

const comicSansTypography = buildTypography(comicSansFontFamily, comicSansFontFamily);
const serifTypography = buildTypography(serifFontFamily, serifFontFamily);

// Playful palette paired with the Comic Sans preset.
const comicSansLightPalette = createPalette({
    mode: 'light',
    primary: {
        main: '#D6336C',
        dark: '#A61E4D',
    },
    secondary: {
        main: '#7048E8',
        dark: '#5F3DC4',
    },
    color: {
        primary: '#1D1B20',
        links: '#7048E8',
        error: '#B44641',
    },
    neutral: {
        primary: '#FFF0F6',
        secondary: '#FFE3EF',
        tertiary: '#FFD6E8',
        quaternary: '#FCC2DC',
        quinary: '#F7A8C9',
    },
    background: {
        paper: '#FFF0F6',
        default: '#FFE3EF',
    },
    low: 'rgb(255, 195, 15)',
    moderate: 'rgb(255, 97, 66)',
    high: 'rgb(205, 0, 117)',
    critical: 'rgb(76, 29, 143)',
});

const comicSansDarkPalette = createPalette({
    mode: 'dark',
    primary: {
        main: '#F06595',
        dark: '#D6336C',
    },
    secondary: {
        main: '#B197FC',
        dark: '#9775FA',
    },
    color: {
        primary: '#FFFFFF',
        links: '#F7A8C9',
        error: '#E9827C',
    },
    neutral: {
        primary: '#1A1015',
        secondary: '#241820',
        tertiary: '#2A1A22',
        quaternary: '#331F29',
        quinary: '#3D2531',
    },
    background: {
        paper: '#2A1A22',
        default: '#1A1015',
    },
    low: 'rgb(255, 195, 15)',
    moderate: 'rgb(255, 97, 66)',
    high: 'rgb(205, 0, 117)',
    critical: 'rgb(76, 29, 143)',
});

// Warm, sepia-toned palette paired with the Serif preset.
const serifLightPalette = createPalette({
    mode: 'light',
    primary: {
        main: '#5F3A1E',
        dark: '#4A2C14',
    },
    secondary: {
        main: '#8C6D3F',
        dark: '#6B5230',
    },
    color: {
        primary: '#2B2017',
        links: '#8C5A2B',
        error: '#B44641',
    },
    neutral: {
        primary: '#FBF6EE',
        secondary: '#F3E9D8',
        tertiary: '#EADDC6',
        quaternary: '#E0CFB2',
        quinary: '#D4BF9B',
    },
    background: {
        paper: '#FBF6EE',
        default: '#F3E9D8',
    },
    low: 'rgb(255, 195, 15)',
    moderate: 'rgb(255, 97, 66)',
    high: 'rgb(205, 0, 117)',
    critical: 'rgb(76, 29, 143)',
});

const serifDarkPalette = createPalette({
    mode: 'dark',
    primary: {
        main: '#C89B6A',
        dark: '#A67C4E',
    },
    secondary: {
        main: '#D9B382',
        dark: '#C89B6A',
    },
    color: {
        primary: '#F5ECE0',
        links: '#E0B07A',
        error: '#E9827C',
    },
    neutral: {
        primary: '#1C1812',
        secondary: '#241F17',
        tertiary: '#2A241C',
        quaternary: '#332C22',
        quinary: '#3D3428',
    },
    background: {
        paper: '#2A241C',
        default: '#1C1812',
    },
    low: 'rgb(255, 195, 15)',
    moderate: 'rgb(255, 97, 66)',
    high: 'rgb(205, 0, 117)',
    critical: 'rgb(76, 29, 143)',
});

// Theme presets bundle a typeface with light/dark palettes. Dark mode continues to toggle between
// the light and dark palette within the selected preset.
export type ThemePreset = 'default' | 'comicSans' | 'serif';

export type ThemePresetConfig = {
    id: ThemePreset;
    label: string;
    typography: Partial<Theme['typography']>;
    lightPalette: Palette;
    darkPalette: Palette;
};

export const themePresets: Record<ThemePreset, ThemePresetConfig> = {
    default: {
        id: 'default',
        label: 'Default',
        typography,
        lightPalette,
        darkPalette,
    },
    comicSans: {
        id: 'comicSans',
        label: 'Comic Sans',
        typography: comicSansTypography,
        lightPalette: comicSansLightPalette,
        darkPalette: comicSansDarkPalette,
    },
    serif: {
        id: 'serif',
        label: 'Serif',
        typography: serifTypography,
        lightPalette: serifLightPalette,
        darkPalette: serifDarkPalette,
    },
};

export const themePresetList: ThemePresetConfig[] = Object.values(themePresets);

export const defaultThemePreset: ThemePreset = 'default';

// Temporary until MUI dialogs migrate to doodle-ui, to fix z-index issue with the side nav bar.
export const themeZIndex: ThemeOptions['zIndex'] = {
    modal: 1410,
};

export const defaultPortalContainer = {
    // Defaults all MUI components that leverage the Modal construct to portal to a child of the applicationContainer element.
    // If not for this, any tailwind based components in a portal and outside the applicationContainer will not respect the current theme.
    // Controlling doodle components: https://tailwindcss.com/docs/dark-mode#toggling-dark-mode-manually
    // Modal construct: https://mui.com/material-ui/api/modal/
    container: () => document.getElementById('app-root'), // Callback so this is re-run on useLayoutEffect within MUI
};

// The word "Label" here is used in the sense of a cypher Label,
// e.g., in the cypher query: `match(u:User) return u`, 'User' is a cypher Label.
// That is to say the "Label" usage here does not reflect a type of AssetGroupTag
export const TagLabelPrefix = 'Tag_' as const;

/**
 * Returns a schema object describing node kinds (`labels`), relationship kinds (`relationshipTypes`),
 * and known property keys. This is primarily used for type completion in the cypher editor.
 *
 * @param kinds - A list of all known kinds in the graph, including:
 *   - Static kinds from Active Directory and Azure
 *   - Dynamically added kinds (e.g., custom types, tier tags, etc.)
 */
export const graphSchema = (kinds: { nodes: string[] | undefined; edges: string[] | undefined }) => {
    // these property keys are not exhaustive as they do not capture potentially generic properties
    const propertyKeys = [
        ...Object.values(CommonKindProperties),
        ...Object.values(ActiveDirectoryKindProperties),
        ...Object.values(AzureKindProperties),
    ];
    const nodeKinds = kinds.nodes ?? [];
    const edgeKinds = kinds.edges ?? [];

    return {
        labels: nodeKinds.map((l) => `:${l}`),
        relationshipTypes: edgeKinds.map((r) => `:${r}`),
        propertyKeys,
    };
};

export const baseGraphLayoutOptions = {
    sequential: 'sequential',
    standard: 'standard',
    table: 'table',
} satisfies MappedStringLiteral<BaseExploreLayoutOptions, BaseExploreLayoutOptions>;

export const baseGraphLayouts = [
    baseGraphLayoutOptions.sequential,
    baseGraphLayoutOptions.standard,
    baseGraphLayoutOptions.table,
] as const;

export const defaultGraphLayout = baseGraphLayoutOptions.sequential;

// Passing these to a router's "future" prop silences noisy warnings from React Router v6
export const reactRouterFutureFlags = {
    v7_relativeSplatPath: true,
    v7_startTransition: true,
};
