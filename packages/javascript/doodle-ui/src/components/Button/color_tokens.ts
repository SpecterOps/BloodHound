import { common, palette } from '../../tailwind/colors';

export const button = {
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
};
