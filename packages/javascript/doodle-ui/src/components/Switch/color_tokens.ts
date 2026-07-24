import { common, palette } from '../../tailwind/colors';

export const switch_classes = {
    fill: { light: palette.neutral.dark[700], dark: palette.white },
    'disabled-fill': { light: palette.neutral.light[300], dark: common.disabled },
    thumb: {
        fill: { light: palette.neutral.light[50], dark: palette.neutral.dark[50] },
        'disabled-fill': { light: palette.neutral.light[400], dark: palette.neutral.light[400] },
    },
};
