import { brand, palette } from '../../tailwind/colors';

// TODO double check colors and class names here
export const input = {
    fill: {
        light: palette.neutral.light[100],
        dark: palette.neutral.dark[400],
    },
    border: {
        default: { light: palette.neutral.dark[700], dark: palette.neutral.light[400] },
        focus: { light: brand.purple.medium, dark: brand.purple.variant },
    },
    //     label: { light: common.dark, dark: common.white },
    //     'fill-disabled': { light: palette.neutral.light[100], dark: palette.neutral.dark[400] },
    //     border: {
    //         hover: { light: light.secondary.main, dark: dark.secondary.main },
    //         disabled: { light: palette.neutral.light[900], dark: palette.neutral.dark[900] },
    //     },
    //     'placeholder-text': { light: text.placeholder, dark: dark.input.placeholder },
};
