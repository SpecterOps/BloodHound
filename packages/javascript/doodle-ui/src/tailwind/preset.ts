// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
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
                primary: 'var(--primary)',
                'primary-variant': 'var(--primary-variant)',

                secondary: 'var(--secondary)',
                'secondary-variant': 'var(--secondary-variant)',
                'secondary-variant-2': 'var(--secondary-variant-2)',

                tertiary: 'var(--tertiary)',
                'tertiary-variant': 'var(--tertiary-variant)',

                contrast: 'var(--contrast)',
                link: 'var(--link)',
                error: 'var(--error)',

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
                'fade-in': {
                    from: {
                        opacity: '0',
                    },
                    to: {
                        opacity: '1',
                    },
                },
                'fade-out': {
                    from: {
                        opacity: '1',
                    },
                    to: {
                        opacity: '0',
                    },
                },
            },
            animation: {
                'accordion-down': 'accordion-down 0.2s ease-out',
                'accordion-up': 'accordion-up 0.2s ease-out',
                'fade-in': 'fade-in 300ms ease-out forwards',
                'fade-out': 'fade-out 300ms ease-out forwards',
            },
            boxShadow: {
                inner1xl: '0px 1px 2px 0px rgba(0, 0, 0, 0.2) inset',
                'outer-1': '0px 1px 2px 0px rgba(0, 0, 0, 0.2)',
                'outer-2': '0px 2px 2px 0px rgba(0, 0, 0, 0.3)',
            },
            fontSize: {
                'headline-1': '2.125rem',
                'headline-2': '1.5rem',
                'headline-3': '1.25rem',
                base: '1rem',
                body: '1rem',
                'body-2': '0.875rem',
                sm: '0.875rem',
                button: '0.875rem',
                caption: '0.75rem',
                eyeline: '0.625rem',
            },
        },
    },
    plugins: [animate],
    safelist: [
        {
            pattern:
                /(bg|text|border|w)-(primary|secondary|accent|contrast|link|error|tertiary|neutral|base|headline|button|caption|eyeline|body|az|group|ou|computer|user|container|meta|domain|default|gpo|aiaca|root-ca|enterprise-ca|nt-auth-store|cert-template|issuance-policy|az-function-app)/,
        },
        {
            pattern: /(w|top|left)-(0|px|[1-9][0-9]?|-[1-9][0-9]?)/, // Safelist padding values including 0, px, and positive/negative numbers
        },
    ],
};
