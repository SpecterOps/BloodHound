import { PluginCreator } from 'tailwindcss/types/config';

const plugin: PluginCreator = ({ addBase, addUtilities }) => {
    addBase({
        ' :root': {
            '--contrast': '#121212',

            '--primary': '#33318f',
            '--primary-variant': '#261f7a',
            '--secondary': '#1a30ff',
            '--secondary-variant': '#0524f0',
            '--secondary-variant-2': '#99a3ff',
            '--tertiary': '#02c577',
            '--tertiary-variant': '#5cc791',

            '--link': '#1a30ff',

            '--neutral-1': '#ffffff',
            '--neutral-2': '#f4f4f4',
            '--neutral-3': '#e3e7ea',
            '--neutral-4': '#dadee1',
            '--neutral-5': '#cacfd3',

            '--error': '#b44641',

            '--neutral-light-1': '#ffffff',
            '--neutral-light-2': '#f4f4f4',
            '--neutral-light-3': '#e3e7ea',
            '--neutral-light-4': '#dadee1',
            '--neutral-light-5': '#cacfd3',

            '--neutral-dark-1': '#121212',
            '--neutral-dark-2': '#222222',
            '--neutral-dark-3': '#272727',
            '--neutral-dark-4': '#2c2c2c',
            '--neutral-dark-5': '#2e2e2e',
        },

        '.dark': {
            '--contrast': '#ffffff',

            '--neutral-1': '#121212',
            '--neutral-2': '#222222',
            '--neutral-3': '#272727',
            '--neutral-4': '#2c2c2c',
            '--neutral-5': '#2e2e2e',

            '--link': '#99a3ff',

            '--error': '#e9827c',
        },
    }),
        addUtilities({
            '.clip-right-rounded': {
                'clip-path': 'inset(0 0.5px 0 -100vw round 0.25rem)',
            },
            '.clip-left-rounded': {
                'clip-path': 'inset(0 -100vw 0 0 round 0.25rem)',
            },
            ".TooltipContent[data-side='top']": {
                'animation-name': 'slideUp',
            },
            ".TooltipContent[data-side='bottom']": {
                'animation-name': 'slideDown',
            },
            '@keyframes slideDown': {
                from: {
                    opacity: '0',
                    transform: 'translateY(-10px)',
                },
                to: {
                    opacity: '1',
                    transform: 'translateY(0)',
                },
            },
            '@keyframes slideUp': {
                from: {
                    opacity: '0',
                    transform: 'translateY(10px)',
                },
                to: {
                    opacity: '1',
                    transform: 'translateY(0)',
                },
            },
        });
};

export default plugin;
