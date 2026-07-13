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
import { PluginCreator } from 'tailwindcss/types/config';
import { common, dark, elevation, light, palette, text } from './colors';

const plugin: PluginCreator = ({ addBase, addUtilities }) => {
    addBase({
        ' :root': {
            // SHARED (same in light and dark)
            // '--common-dark': common.dark,
            // '--common-white': common.white,

            // // UTILITIES / risk level
            // '--risk-critical': palette.purple.A300,
            // '--risk-high': palette.red.A300,
            // '--risk-moderate': palette.brown[300],
            // '--risk-low': palette.yellow.A300,
            // '--risk-mitigated': palette.green.A300,
            // '--risk-resolved': palette['light-blue'].A300,
            // '--risk-accepted': palette.neutral.light[300],
            // '--risk-text': common.black,

            // BRAND
            // '--brand-primary-highlight-neon-blue': common['neon-blue'],
            // '--brand-primary-highlight-orange': common.orange,
            // '--brand-secondary-highlight-light-blue': light.secondary.main,
            // '--brand-secondary-medium-purple': common['medium-purple'],
            // '--brand-secondary-light-purple': common['light-purple'],
            // '--brand-secondary-light-blue-gray': common['light-blue-gray'],

            // END OF SHARED

            // MAIN
            '--primary': light.primary.main,
            '--primary-variant': light.primary.variant,
            '--secondary-main': light.secondary.main,
            '--secondary-variant': light.secondary.variant,
            // '--tertiary': light.tertiary.main,
            // '--tertiary-variant': light.tertiary.variant,
            '--disabled': light.disabled,

            // // NEUTRALS
            '--neutral-50': palette.neutral.light[50],
            // '--neutral-100': palette.neutral.light[100],
            '--neutral-200': palette.neutral.light[200],
            '--neutral-300': palette.neutral.light[300],
            '--neutral-400': palette.neutral.light[400],
            // '--neutral-500': palette.neutral.light[500],
            // '--neutral-600': palette.neutral.light[600],
            // '--neutral-700': palette.neutral.light[700],
            // '--neutral-800': palette.neutral.light[800],
            // '--neutral-900': palette.neutral.light[900],

            // // TEXT
            '--text-main': common.dark,
            '--text-light': text.light,
            '--text-contrast': common.white,
            '--text-disabled': light.text.disabled,
            // '--text-primary': light.primary.main,
            // '--text-secondary': light.secondary.main,

            // LINKS
            '--link-main': light.secondary.main,
            '--link-hover': light.secondary.variant,

            // // ELEVATION
            // '--elevation-0': elevation.light[0],
            // '--elevation-1': elevation.light[1],
            // '--elevation-2': elevation.light[2],
            // '--elevation-3': elevation.light[3],
            // '--elevation-4': elevation.light[4],
            // '--elevation-5': elevation.light[5],

            // STATUS
            '--status-error-main': light.status.error.main,
            '--status-error-text': light.status.error.text,
            '--status-error-fill': light.status.error.fill,
            '--status-warning-main': light.status.warning.main,
            '--status-warning-text': light.status.warning.text,
            '--status-warning-fill': light.status.warning.fill,
            '--status-success-main': light.status.success.main,
            '--status-success-text': light.status.success.text,
            '--status-success-fill': light.status.success.fill,
            '--status-info-main': light.status.info.main,
            '--status-info-text': light.status.info.text,
            '--status-info-fill': light.status.info.fill,
            '--status-indeterminate-fill': light.status.indeterminate.fill,

            // // BRAND COLORS
            '--bhe-main': light.primary.main,
            // '--sp-main': light.primary.main,
            '--bhce-main': light['bhce-main'],
            // '--logo-neutral': common.black,
            // '--brand-primary-dark-purple': light.primary.variant,
            // '--brand-primary-deep-purple': light.primary.main,
            // '--brand-primary-medium-blue': light.secondary.main,
            // '--brand-secondary-highlight-green': light.tertiary.main,

            // // Components/Button
            // '--secondary-btn-fill': palette.neutral.light[300],
            // '--secondary-btn-active-fill': palette.neutral.light[500],
            // '--btn-disabled-fill': palette.neutral.light[200],
            // '--toggle-btn-fill': common.white,
            // '--toggle-btn-border': palette.neutral.light[500],

            // // Components/Input
            // '--input-label': common.dark,
            '--input-fill': elevation.light[1],
            // '--input-fill-disabled': palette.neutral.light[100],
            '--input-border-default': palette.grey[700],
            // '--input-border-hover': light.secondary.main,
            // '--input-border-disabled': palette.neutral.light[900],
            // '--input-placeholder-text': text.placeholder,

            // // Components/Input/Selectors
            // '--selector-disable-fill': common.white,
            '--switch-fill': palette.neutral.dark[700],
            '--switch-disabled-fill': palette.neutral.light[300],

            // // Components/Data Display/ Menu
            // '--menu-bg': elevation.light[0],

            // // Components/Data Display/ Badge & Chip
            '--badge-primary-fill': light.badge.primary.fill,
            '--badge-primary-outline': light.badge.primary.outline,

            '--badge-secondary-fill': light.badge.secondary.fill,
            '--badge-secondary-outline': light.badge.secondary.outline,

            '--badge-grey-fill': light.badge.grey.fill,
            '--badge-grey-outline': light.badge.grey.outline,

            '--badge-red-fill': light.badge.red.fill,
            '--badge-red-outline': light.badge.red.outline,

            '--badge-orange-fill': light.badge.orange.fill,
            '--badge-orange-outline': light.badge.orange.outline,

            '--badge-green-fill': light.badge.green.fill,
            '--badge-green-outline': light.badge.green.outline,

            '--badge-blue-fill': light.badge.blue.fill,
            '--badge-blue-outline': light.badge.blue.outline,

            // // Components/Data Display/Badge & Chip / Chip
            // '--chip-indeterminate': palette.neutral.light[300],
            // '--chip-indeterminate-hover': palette.neutral.light[600],
            // '--chip-outline-fill': common.white,
            // '--chip-outline-hover': palette.neutral.light[100],
            // '--chip-outline-active': palette.neutral.light[300],

            // // Components/ SideNav
            // '--sidenav-bg': palette.neutral.light[50],
            // '--sidenav-bg-hover': palette.neutral.light[200],
            // '--sidenav-bg-active': palette.neutral.light[300],
            // '--nav-control-btn': palette.neutral.light[300],
            // '--nav-control-btn-hover': palette.neutral.light[600],
            // '--nav-control-btn-focus': palette.neutral.light[500],

            // // Components/ SideNav/ Nav items
            // '--nav-item-default': common.dark,
            // '--nav-item-hover': light.secondary.main,
            // '--nav-item-active': light.primary.main,

            // // Components / Graph
            // '--platform-badge-bg': elevation.light[2],

            // // Components / Graph / Edge
            // '--edge-color': light.edge.color,
            // '--edge-label-bg': palette.neutral.light[200],
            // '--edge-label-text': common.dark,

            // // Utilities
            // '--icon': common.dark,
            // '--icon-contrast': common.white,
            // '--icon-disabled': palette.grey[700],
            // '--divider': palette.neutral.light[500],

            // Legacy below
            // same as palette.neutral.dark[50]
            '--contrast': '#121212',

            '--secondary-variant-2': '#99a3ff',
            '--tertiary': '#02c577',
            '--tertiary-variant': '#5cc791',

            // same as common.white
            '--neutral-1': '#ffffff',
            // same as neutral-100
            '--neutral-2': '#f4f4f4',
            '--neutral-3': '#e3e7ea',
            '--neutral-4': '#dadee1',
            '--neutral-5': '#cacfd3',

            '--link': '#1a30ff',

            '--error': '#b44641',

            // same as common.white
            '--neutral-light-1': '#ffffff',
            // same as neutral-100
            '--neutral-light-2': '#f4f4f4',
            '--neutral-light-3': '#e3e7ea',
            '--neutral-light-4': '#dadee1',
            // same as neutral-light-400
            '--neutral-light-5': '#cacfd3',

            // same palette.neutral.dark[50]
            '--neutral-dark-1': '#121212',
            '--neutral-dark-2': '#222222',
            '--neutral-dark-3': '#272727',
            '--neutral-dark-4': '#2c2c2c',
            // same as palette.neutral.dark[700]
            '--neutral-dark-5': '#2e2e2e',
        },

        '.dark': {
            // MAIN
            '--primary': dark.primary.main,
            '--primary-variant': dark.primary.variant,
            '--secondary': dark.secondary.main,
            '--secondary-variant': dark.secondary.variant,
            // '--tertiary': dark.tertiary.main,
            // '--tertiary-variant': dark.tertiary.variant,
            '--disabled': dark.disabled,

            // // NEUTRALS
            '--neutral-50': palette.neutral.dark[50],
            // '--neutral-100': palette.neutral.dark[100],
            // '--neutral-200': palette.neutral.dark[200],
            // '--neutral-300': palette.neutral.dark[300],
            // '--neutral-400': palette.neutral.dark[400],
            // '--neutral-500': palette.neutral.dark[500],
            // '--neutral-600': palette.neutral.dark[600],
            // '--neutral-700': palette.neutral.dark[700],
            // '--neutral-800': palette.neutral.dark[800],
            // '--neutral-900': palette.neutral.dark[900],

            // // TEXT
            '--text-main': common.white,
            '--text-light': text.dark,
            '--text-contrast': common.dark,
            '--text-disabled': common.disabled,
            // '--text-primary': dark.primary.main,
            // '--text-secondary': dark.secondary.main,

            // LINKS
            '--link-main': dark.secondary.main,
            '--link-hover': dark.secondary.variant,

            // // ELEVATION
            // '--elevation-0': elevation.dark[0],
            '--elevation-1': elevation.dark[1],
            // '--elevation-2': elevation.dark[2],
            // '--elevation-3': elevation.dark[3],
            // '--elevation-4': elevation.dark[4],
            // '--elevation-5': elevation.dark[5],

            // STATUS
            '--status-error-main': dark.status.error.main,
            '--status-error-text': dark.status.error.text,
            '--status-error-fill': dark.status.error.fill,
            '--status-warning-main': dark.status.warning.main,
            '--status-warning-text': dark.status.warning.text,
            '--status-warning-fill': dark.status.warning.fill,
            '--status-success-main': dark.status.success.main,
            '--status-success-text': dark.status.success.text,
            '--status-success-fill': dark.status.success.fill,
            '--status-info-main': dark.status.info.main,
            '--status-info-text': dark.status.info.text,
            '--status-info-fill': dark.status.info.fill,
            '--status-indeterminate-fill': dark.status.indeterminate.fill,

            // // BRAND COLORS
            '--bhe-main': dark.primary.main,
            // '--sp-main': dark.primary.main,
            '--bhce-main': dark['bhce-main'],
            // '--logo-neutral': common.white,
            // '--brand-primary-dark-purple': dark.primary.variant,
            // '--brand-primary-deep-purple': dark.primary.main,
            // '--brand-primary-medium-blue': dark.secondary.main,
            // '--brand-secondary-highlight-green': dark.tertiary.main,

            // // Components/Button
            // '--secondary-btn-fill': palette.neutral.dark[800],
            // '--secondary-btn-active-fill': palette.neutral.dark[600],
            // '--btn-disabled-fill': palette.neutral.dark[700],
            // '--toggle-btn-fill': common.dark,
            // '--toggle-btn-border': palette.neutral.dark[600],

            // // Components/Input
            // '--input-label': common.white,
            '--input-fill': elevation.dark[1],
            // '--input-fill-disabled': palette.neutral.dark[400],
            '--input-border-default': dark.input.border,
            // '--input-border-hover': dark.secondary.main,
            // '--input-border-disabled': palette.neutral.dark[900],
            // '--input-placeholder-text': dark.input.placeholder,

            // // Components/Input/Selectors
            // '--selector-disable-fill': common.white,
            '--switch-fill': common.white,
            '--switch-disabled-fill': common.disabled,

            // // Components/Data Display/ Menu
            // '--menu-bg': elevation.dark[2],

            // // Components/Data Display/ Badge & Chip
            '--badge-primary-fill': dark.badge.primary.fill,
            '--badge-primary-outline': dark.badge.primary.outline,

            '--badge-secondary-fill': dark.badge.secondary.fill,
            '--badge-secondary-outline': dark.badge.secondary.outline,

            '--badge-grey-fill': dark.badge.grey.fill,
            '--badge-grey-outline': dark.badge.grey.outline,

            '--badge-red-fill': dark.badge.red.fill,
            '--badge-red-outline': dark.badge.red.outline,

            '--badge-orange-fill': dark.badge.orange.fill,
            '--badge-orange-outline': dark.badge.orange.outline,

            '--badge-green-fill': dark.badge.green.fill,
            '--badge-green-outline': dark.badge.green.outline,

            '--badge-blue-fill': dark.badge.blue.fill,
            '--badge-blue-outline': dark.badge.blue.outline,

            // // Components/Data Display/Badge & Chip / Chip
            // '--chip-indeterminate': palette.neutral.dark[500],
            // '--chip-indeterminate-hover': palette.neutral.dark[300],
            // '--chip-outline-fill': elevation.dark[1],
            // '--chip-outline-hover': palette.neutral.dark[200],
            // '--chip-outline-active': palette.neutral.dark[400],

            // // Components/ SideNav
            // '--sidenav-bg': palette.neutral.dark[300],
            // '--sidenav-bg-hover': palette.neutral.dark[200],
            // '--sidenav-bg-active': palette.neutral.dark[600],
            // '--nav-control-btn': palette.neutral.dark[300],
            // '--nav-control-btn-hover': palette.neutral.dark[500],
            // '--nav-control-btn-focus': palette.neutral.dark[400],

            // // Components/ SideNav/ Nav items
            // '--nav-item-default': common.white,
            // '--nav-item-hover': dark.secondary.main,
            // '--nav-item-active': dark.primary.main,

            // // Components / Graph
            // '--platform-badge-bg': elevation.dark[2],

            // // Components / Graph / Edge
            // '--edge-color': dark.edge.color,
            // '--edge-label-bg': palette.neutral.dark[500],
            // '--edge-label-text': common.white,

            // // Utilities
            // '--icon': common.white,
            // '--icon-contrast': common.dark,
            // '--icon-disabled': common.disabled,
            // '--divider': palette.neutral.dark[500],

            ///////////// Legacy below //////////
            // same as common.white
            '--contrast': '#ffffff',
            // same palette.neutral.dark[50]
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
