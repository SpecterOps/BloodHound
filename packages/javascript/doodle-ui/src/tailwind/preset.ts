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
                '3xl': '1920px',
            },
        },
        extend: {
            colors: {
                primary: 'var(--primary)',
                'primary-variant': 'var(--primary-variant)',

                secondary: 'var(--secondary-main)',
                'secondary-variant': 'var(--secondary-variant)',
                'secondary-variant-2': 'var(--secondary-variant-2)',

                tertiary: 'var(--tertiary)',
                'tertiary-variant': 'var(--tertiary-variant)',

                'text-main': 'var(--text-main)',
                'text-disabled': 'var(--text-disabled)',

                disabled: 'var(--disabled)',
                contrast: 'var(--contrast)',
                link: 'var(--link-main)',
                'link-hover': '(--link-hover)',
                error: 'var(--error)',

                'switch-fill': 'var(--switch-fill)',
                'switch-disabled-fill': 'var(--switch-disabled-fill)',

                'neutral-50': 'var(--neutral-50)',
                'neutral-400': 'var(--neutral-400)',

                // Legacy
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

                // New color tokens:

                // tertiary: 'var(--tertiary)',
                // 'tertiary-variant': 'var(--tertiary-variant)',

                // 'common-dark': 'var(--common-dark)',
                // 'common-white': 'var(--common-white)',

                // contrast: 'var(--common-dark)',

                'text-light': 'var(--text-light)',
                // 'text-contrast': 'var(--text-contrast)',
                // 'text-primary': 'var(--text-primary)',
                // 'text-secondary': 'var(--text-secondary)',

                // 'neutral-100': 'var(--neutral-100)',
                'neutral-200': 'var(--neutral-200)',
                // 'neutral-300': 'var(--neutral-300)',
                // 'neutral-400': 'var(--neutral-400)',
                // 'neutral-500': 'var(--neutral-500)',
                // 'neutral-600': 'var(--neutral-600)',
                // 'neutral-700': 'var(--neutral-700)',
                // 'neutral-800': 'var(--neutral-800)',
                // 'neutral-900': 'var(--neutral-900)',

                'status-error-main': 'var(--status-error-main)',
                'status-error-text': 'var(--status-error-text)',
                'status-error-fill': 'var(--status-error-fill)',
                'status-warning-main': 'var(--status-warning-main)',
                'status-warning-text': 'var(--status-warning-text)',
                'status-warning-fill': 'var(--status-warning-fill)',
                'status-success-main': 'var(--status-success-main)',
                'status-success-text': 'var(--status-success-text)',
                'status-success-fill': 'var(--status-success-fill)',
                'status-info-main': 'var(--status-info-main)',
                'status-info-text': 'var(--status-info-text)',
                'status-info-fill': 'var(--status-info-fill)',
                'status-indeterminate-fill': 'var(--status-indeterminate-fill)',

                // 'bhe-main': 'var(--bhe-main)',
                // 'sp-main': 'var(--sp-main)',
                // 'bhce-main': 'var(--bhce-main)',
                // 'logo-neutral': 'var(--logo-neutral)',

                // 'brand-primary-dark-purple': 'var(--brand-primary-dark-purple)',
                // 'brand-primary-deep-purple': 'var(--brand-primary-deep-purple)',
                // 'brand-primary-medium-blue': 'var(--brand-primary-medium-blue)',
                // 'brand-primary-highlight-neon-blue': 'var(--brand-primary-highlight-neon-blue)',
                // 'brand-primary-highlight-orange': 'var(--brand-primary-highlight-orange)',
                // 'brand-secondary-medium-purple': 'var(--brand-secondary-medium-purple)',
                // 'brand-secondary-light-purple': 'var(--brand-secondary-light-purple)',
                // 'brand-secondary-light-blue-gray': 'var(--brand-secondary-light-blue-gray)',
                // 'brand-secondary-highlight-light-blue': 'var(--brand-secondary-highlight-light-blue)',
                // 'brand-secondary-highlight-green': 'var(--brand-secondary-highlight-green)',

                // 'secondary-btn-fill': 'var(--secondary-btn-fill)',
                // 'secondary-btn-active-fill': 'var(--secondary-btn-active-fill)',
                // 'btn-disabled-fill': 'var(--btn-disabled-fill)',
                // 'toggle-btn-fill': 'var(--toggle-btn-fill)',
                // 'toggle-btn-border': 'var(--toggle-btn-border)',

                // 'input-label': 'var(--input-label)',
                // 'input-fill': 'var(--input-fill)',
                // 'input-fill-disabled': 'var(--input-fill-disabled)',
                // 'input-border-default': 'var(--input-border-default)',
                // 'input-border-hover': 'var(--input-border-hover)',
                // 'input-border-disabled': 'var(--input-border-disabled)',
                // 'input-placeholder-text': 'var(--input-placeholder-text)',

                // 'selector-disable-fill': 'var(--selector-disable-fill)',

                // 'menu-bg': 'var(--menu-bg)',

                'badge-primary-fill': 'var(--badge-primary-fill)',
                'badge-primary-outline': 'var(--badge-primary-outline)',

                'badge-secondary-fill': 'var(--badge-secondary-fill)',
                'badge-secondary-outline': 'var(--badge-secondary-outline)',

                'badge-grey-fill': 'var(--badge-grey-fill)',
                'badge-grey-outline': 'var(--badge-grey-outline)',

                'badge-red-fill': 'var(--badge-red-fill)',
                'badge-red-outline': 'var(--badge-red-outline)',

                'badge-orange-fill': 'var(--badge-orange-fill)',
                'badge-orange-outline': 'var(--badge-orange-outline)',

                'badge-green-fill': 'var(--badge-green-fill)',
                'badge-green-outline': 'var(--badge-green-outline)',

                'badge-blue-fill': 'var(--badge-blue-fill)',
                'badge-blue-outline': 'var(--badge-blue-outline)',

                // 'sidenav-bg': 'var(--sidenav-bg)',
                // 'sidenav-bg-hover': 'var(--sidenav-bg-hover)',
                // 'sidenav-bg-active': 'var(--sidenav-bg-active)',
                // 'nav-control-btn': 'var(--nav-control-btn)',
                // 'nav-control-btn-hover': 'var(--nav-control-btn-hover)',
                // 'nav-control-btn-focus': 'var(--nav-control-btn-focus)',
                // 'nav-item-default': 'var(--nav-item-default)',
                // 'nav-item-hover': 'var(--nav-item-hover)',
                // 'nav-item-active': 'var(--nav-item-active)',

                // 'platform-badge-bg': 'var(--platform-badge-bg)',

                // 'edge-color': 'var(--edge-color)',
                // 'edge-label-bg': 'var(--edge-label-bg)',
                // 'edge-label-text': 'var(--edge-label-text)',

                // icon: 'var(--icon)',
                // 'icon-contrast': 'var(--icon-contrast)',
                // 'icon-disabled': 'var(--icon-disabled)',
                // divider: 'var(--divider)',

                // 'risk-critical': 'var(--risk-critical)',
                // 'risk-high': 'var(--risk-high)',
                // 'risk-moderate': 'var(--risk-moderate)',
                // 'risk-low': 'var(--risk-low)',
                // 'risk-mitigated': 'var(--risk-mitigated)',
                // 'risk-resolved': 'var(--risk-resolved)',
                // 'risk-accepted': 'var(--risk-accepted)',
                // 'risk-text': 'var(--risk-text)',
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
                /(bg|text|border|w|ring|ring-offset)-(primary|secondary|accent|contrast|link|error|tertiary|neutral|base|headline|button|caption|eyeline|body|az|group|ou|computer|user|container|meta|domain|default|gpo|aiaca|root-ca|enterprise-ca|nt-auth-store|cert-template|issuance-policy|az-function-app)/,
        },
        {
            pattern: /(w|top|left)-(0|px|[1-9][0-9]?|-[1-9][0-9]?)/, // Safelist padding values including 0, px, and positive/negative numbers
        },
    ],
};
