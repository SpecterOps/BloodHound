// Copyright 2023 Specter Ops, Inc.
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

import { Theme } from '@mui/material';
import { addOpacityToHex } from './utils/colors';

export const NODE_GRAPH_RENDER_LIMIT = 1000;

export const ZERO_VALUE_API_DATE = '0001-01-01T00:00:00Z';

export const TIER_ZERO_TAG = 'admin_tier_0';
export const TIER_ZERO_LABEL = 'High Value';
export const OWNED_OBJECT_TAG = 'owned';

export const lightPalette = {
    primary: {
        main: '#33318F',
        dark: '#261F7A',
    },
    secondary: {
        main: '#1A30FF',
        dark: '#0524F0',
    },
    tertiary: {
        main: '#5CC791',
        dark: '#02C577',
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
};

export const darkPalette = {
    primary: {
        main: '#33318F',
        dark: '#261F7A',
    },
    secondary: {
        main: '#1A30FF',
        dark: '#0524F0',
    },
    tertiary: {
        main: '#5CC791',
        dark: '#02C577',
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
};

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

export const components = (theme: Theme): Partial<Theme['components']> => ({
    MuiAccordionSummary: {
        styleOverrides: {
            root: {
                flexDirection: 'row-reverse',
            },
            content: {
                marginRight: '4px',
            },
        },
    },
    MuiLink: {
        styleOverrides: {
            root: {
                color: theme.palette.color.links,
            },
        },
    },
    MuiInputLabel: {
        styleOverrides: {
            root: {
                '&.Mui-focused': {
                    color: theme.palette.color.links,
                },
            },
        },
    },
    MuiTextField: {
        styleOverrides: {
            root: {
                '&:hover .MuiInputBase-root .MuiOutlinedInput-notchedOutline': {
                    borderColor: theme.palette.color.links,
                },
                '& .MuiInputBase-root.Mui-focused .MuiOutlinedInput-notchedOutline': {
                    borderColor: theme.palette.color.links,
                },
            },
        },
    },
    MuiInput: {
        styleOverrides: {
            underline: {
                '&:after': {
                    borderBottom: `2px solid ${theme.palette.color.links}`,
                },
                '&:hover:not($disabled):not($focused):not($error):before': {
                    borderBottom: `2px solid ${theme.palette.color.links}`,
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
                    backgroundColor: theme.palette.neutral.secondary,
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
    },
    MuiDialogActions: {
        styleOverrides: {
            root: {
                padding: theme.spacing(2, 3),
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
                '& svg': {
                    color: theme.palette.color.primary,
                },
            },
        },
    },
    MuiTabs: {
        styleOverrides: {
            root: {
                '& .MuiTab-labelIcon': {
                    color: theme.palette.color.links,
                },
                '& .MuiButtonBase-root.Mui-selected': {
                    color: theme.palette.color.links,
                },
                '& .MuiTab-labelIcon:not(.Mui-selected)': {
                    color: theme.palette.color.primary,
                },
                '& .MuiTabs-indicator': {
                    backgroundColor: theme.palette.color.links,
                },
                '& .Mui-selected > svg': {
                    color: theme.palette.color.links,
                },
                '& :not(.Mui-selected) > svg': {
                    color: theme.palette.color.primary,
                },
            },
        },
    },
    MuiAlert: {
        styleOverrides: {
            root: {
                '&.MuiAlert-standardWarning': {
                    backgroundColor: addOpacityToHex(theme.palette.warning.main, 20),
                },
                '&.MuiAlert-standardInfo': {
                    backgroundColor: addOpacityToHex(theme.palette.info.main, 20),
                },
                '&.MuiAlert-standardError': {
                    backgroundColor: addOpacityToHex(theme.palette.error.main, 20),
                },
            },
        },
    },
    MuiLinearProgress: {
        styleOverrides: {
            root: {
                backgroundColor: addOpacityToHex(theme.palette.primary.main, 40),
                '& .MuiLinearProgress-barColorPrimary': {
                    backgroundColor: theme.palette.primary.main,
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
