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
export const MAX_NAME_LENGTH = 1000;
export const MIN_NAME_LENGTH = 2;
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

export const typography: Partial<Theme['typography']> = {
    h1: {
        fontWeight: 400,
        fontSize: '1.8rem',
        lineHeight: 2,
        letterSpacing: 0,
    },
    h2: {
        fontWeight: 500,
        fontSize: '1.5rem',
        lineHeight: 1.5,
        letterSpacing: 0,
    },
    h3: {
        fontWeight: 500,
        fontSize: '1.2rem',
        lineHeight: 1.25,
        letterSpacing: 0,
    },
    h4: {
        fontWeight: 500,
        fontSize: '1.25rem',
        lineHeight: 1.5,
        letterSpacing: 0,
    },
    h5: {
        fontWeight: 700,
        fontSize: '1.125rem',
        lineHeight: 1.5,
        letterSpacing: 0.25,
    },
    h6: {
        fontWeight: 700,
        fontSize: '1.0rem',
        lineHeight: 1.5,
        letterSpacing: 0.25,
    },
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
