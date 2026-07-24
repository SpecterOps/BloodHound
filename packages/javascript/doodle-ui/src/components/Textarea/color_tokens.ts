import { brand, common, palette } from '../../tailwind/colors';

export const textarea = {
    fill: { light: common.white, dark: palette.neutral.dark[700] },
    border: {
        default: { light: palette.neutral.light[400], dark: palette.neutral.dark[700] },
        hover: { light: brand.purple.medium, dark: brand.purple.variant },
    },
};
