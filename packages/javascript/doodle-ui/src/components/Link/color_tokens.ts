import { brand } from '../../tailwind/colors';

export const link = {
    DEFAULT: {
        light: brand.purple.medium,
        dark: brand.blue.medium,
        varName: 'link-main',
    },
    hover: {
        light: '#3729BB',
        dark: '#4E95FF',
    },
    legacy: {
        light: '#1a30ff',
        dark: brand.purple.variant,
        cssVariableOnly: true,
        varName: 'link',
    },
};
