import { common, palette } from '../../../../doodle-ui/src/tailwind/colors';

export const dropdown = {
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
};
