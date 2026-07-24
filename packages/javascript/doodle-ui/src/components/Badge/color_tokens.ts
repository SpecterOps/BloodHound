import { brand, palette } from '../../tailwind/colors';

export const badge = {
    light: {
        primary: {
            fill: brand.purple.light,
            outline: brand.purple.dark,
        },
        secondary: {
            fill: brand.blue.medium,
            outline: brand.purple.medium,
        },
        grey: {
            fill: palette.neutral.light[200],
            outline: palette.neutral.light[400],
        },
        red: {
            // same as status.error.light
            fill: palette.red[200],
            // same as status.error.medium or same as theme.red.medium
            outline: '#B44641',
        },
        orange: {
            // same as status.warning.light
            fill: palette.orange[200],
            // same as status.warning.medium
            outline: palette.orange[900],
        },
        green: {
            // same as status.success.light
            fill: palette.green[200],
            // same as status.success.medium
            outline: palette.green[600],
        },
        blue: {
            fill: brand.blue.light,
            outline: brand.blue.dark,
        },
    },
    dark: {
        primary: {
            fill: '#605DF7',
            outline: brand.purple.light,
        },
        secondary: {
            fill: '#1569E7',
            outline: brand.blue.medium,
        },
        grey: {
            fill: palette.neutral.dark[500],
            outline: palette.neutral.dark[700],
        },
        red: {
            fill: palette.red[700],
            // same as status.error.light
            outline: palette.red[200],
        },
        orange: {
            fill: '#C15012',
            // same as status.warning.light
            outline: palette.orange[200],
        },
        green: {
            fill: palette.green[800],
            // same as status.success.light
            outline: palette.green[200],
        },
        blue: {
            fill: palette['light-blue'][800],
            outline: brand.blue.light,
        },
    },
};
