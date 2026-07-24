import { brand, palette } from '../../tailwind/colors';

export const radio = {
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
};
