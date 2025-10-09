import animate from 'tailwindcss-animate';

export default {
    theme: {
        fontFamily: {
            sans: ['Roboto', 'Helvetica', 'Arial', 'sans-serif'],
        },
        container: {
            center: true,
            padding: '2rem',
            screens: {
                '2xl': '1400px',
            },
        },
        extend: {
            colors: {
                contrast: 'var(--contrast)',

                primary: 'var(--primary)',
                'primary-variant': 'var(--primary-variant)',

                secondary: 'var(--secondary)',
                'secondary-variant': 'var(--secondary-variant)',
                'secondary-variant-2': 'var(--secondary-variant-2)',

                link: 'var(--link)',
                error: 'var(--error)',

                tertiary: 'var(--tertiary)',
                'tertiary-variant': 'var(--tertiary-variant)',

                'neutral-1': 'var(--neutral-1)',
                'neutral-2': 'var(--neutral-2)',
                'neutral-3': 'var(--neutral-3)',
                'neutral-4': 'var(--neutral-4)',
                'neutral-5': 'var(--neutral-5)',

                'neutral-light-1': 'var(--neutral-light-1)',
                'neutral-light-2': 'var(--neutral-light-2)',
                'neutral-light-3': 'var(--neutral-light-3)',
                'neutral-light-4': 'var(--neutral-light-4)',
                'neutral-light-5': 'var(--neutral-light-5)',

                'neutral-dark-1': 'var(--neutral-dark-1)',
                'neutral-dark-2': 'var(--neutral-dark-2)',
                'neutral-dark-3': 'var(--neutral-dark-3)',
                'neutral-dark-4': 'var(--neutral-dark-4)',
                'neutral-dark-5': 'var(--neutral-dark-5)',
            },
            keyframes: {
                'accordion-down': {
                    from: { height: '0' },
                    to: { height: 'var(--radix-accordion-content-height)' },
                },
                'accordion-up': {
                    from: { height: 'var(--radix-accordion-content-height)' },
                    to: { height: '0' },
                },
            },
            animation: {
                'accordion-down': 'accordion-down 0.2s ease-out',
                'accordion-up': 'accordion-up 0.2s ease-out',
            },
            boxShadow: {
                inner1xl: '0px 1px 2px 0px rgba(0, 0, 0, 0.2) inset',
                'outer-1': '0px 1px 2px 0px rgba(0, 0, 0, 0.2)',
                'outer-2': '0px 2px 2px 0px rgba(0, 0, 0, 0.3)',
            },
        },
    },
    plugins: [animate],
};
