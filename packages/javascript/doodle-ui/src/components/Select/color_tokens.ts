import { brand, common, palette } from '../../tailwind/colors';

export const select = {
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
};
