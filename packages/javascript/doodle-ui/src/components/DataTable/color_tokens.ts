import { brand, palette } from '../../tailwind/colors';

export const data_table_colors = {
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
};
