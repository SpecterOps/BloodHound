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

const secondaryVariant2 = '#99a3ff';
const darkDataTableRowSelectedOutline = '#4A42B5';
const focusRingWidth = '2px';
const focusRingOffsetWidth = '2px';

const plugin: PluginCreator = ({ addBase, addUtilities }) => {
    addBase({
        ' :root': {
            // SHARED (same in light and dark)
            '--common-dark': common.dark,
            '--common-white': common.white,

            // UTILITIES / risk level
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
            '--secondary': light.secondary.main,
            '--secondary-variant': light.secondary.variant,
            '--tertiary': light.tertiary.main,
            '--tertiary-variant': light.tertiary.variant,
            '--disabled': light.disabled,

            // Shared control and surface tokens
            '--border-control': palette.grey[700],
            '--control-accent': 'var(--secondary)',
            '--surface': common.white,
            '--surface-subtle': palette.neutral.light[100],
            '--surface-hover': palette.neutral.light[300],

            // NEUTRALS
            '--neutral-50': palette.neutral.light[50],
            '--neutral-100': palette.neutral.light[100],
            '--neutral-200': palette.neutral.light[200],
            '--neutral-300': palette.neutral.light[300],
            '--neutral-400': palette.neutral.light[400],
            '--neutral-500': palette.neutral.light[500],
            '--neutral-600': palette.neutral.light[600],
            '--neutral-700': palette.neutral.light[700],
            '--neutral-800': palette.neutral.light[800],
            '--neutral-900': palette.neutral.light[900],

            // TEXT
            '--text-main': common.dark,
            '--text-light': text.light,
            '--text-contrast': common.white,
            '--text-faint': text.faint,

            // LINKS
            '--link-main': 'var(--secondary)',
            '--link-hover': 'var(--secondary-variant)',

            // FOCUS
            '--focus-ring': 'var(--secondary)',
            '--focus-ring-offset': common.white,
            '--focus-ring-width': focusRingWidth,
            '--focus-ring-offset-width': focusRingOffsetWidth,

            // ELEVATION
            '--elevation-0': elevation.light[0],
            '--elevation-1': elevation.light[1],
            '--elevation-2': elevation.light[2],
            '--elevation-3': elevation.light[3],
            '--elevation-4': elevation.light[4],
            '--elevation-5': elevation.light[5],

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

            // BRAND COLORS
            '--bhe-main': 'var(--primary)', // TODO BED-10220 is this necessary?
            // '--sp-main': 'var(--primary)',
            '--bhce-main': light['bhce-main'],
            // '--logo-neutral': common.black,
            // '--brand-primary-dark-purple': 'var(--primary-variant)',
            // '--brand-primary-deep-purple': 'var(--primary)',
            // '--brand-primary-medium-blue': 'var(--secondary)',
            // '--brand-secondary-highlight-green': 'var(--tertiary)',

            // Components/Accordion
            '--accordion-fill': 'var(--neutral-light-2)',

            // Components/Card
            '--card-border': 'var(--neutral-light-4)',

            // Components/Table
            '--table-fill': 'var(--neutral-light-2)',
            '--table-row-border-width': '1px',
            '--table-dragging-fill': 'var(--neutral-light-1)',

            // Components/Tabs
            '--tabs-text': 'inherit',
            '--tabs-active-accent': 'var(--primary)',

            // Components/Typography
            '--typography-body-text': 'var(--text-light)',

            // Components/Breadcrumbs
            '--breadcrumb-text': 'inherit',
            '--breadcrumb-page-text': 'var(--primary)',

            // Components/DatePicker
            '--date-picker-input-fill': 'var(--common-white)',
            '--date-picker-content-border': 'var(--neutral-light-5)',

            // Components/Pagination
            '--pagination-text': 'inherit',
            '--pagination-disabled-text': palette.neutral.light[800],

            // Components/Skeleton
            '--skeleton-fill': 'var(--neutral-light-5)',

            // Components/ComboBadge
            '--combo-badge-fill': 'var(--neutral-light-5)',
            '--combo-badge-text': 'var(--primary)',

            // Components/Dialog
            '--dialog-overlay-fill': 'rgb(0 0 0 / 0.4)',
            '--dialog-fill': 'var(--neutral-light-2)',
            '--dialog-text': 'inherit',

            // Components/Popover
            '--popover-fill': 'var(--neutral-light-2)',
            '--popover-text': 'var(--neutral-dark-1)',

            // Components/Slider
            '--slider-track-fill': 'var(--neutral-200)',
            '--slider-thumb-fill': 'var(--neutral-100)',
            '--slider-thumb-border': 'var(--neutral-300)',
            '--slider-thumb-focus-fill': 'var(--neutral-100)',
            '--slider-thumb-active-fill': 'var(--neutral-100)',

            // Components/Stepper
            '--stepper-inactive-border': '#939597',

            // Components/Toggle
            '--toggle-btn-fill': 'var(--surface)',
            '--toggle-btn-border': palette.neutral.light[500],
            '--toggle-group-fill': 'var(--surface-subtle)',

            '--toggle-hover-border': 'var(--secondary)',
            '--toggle-active-border': 'var(--primary-variant)',
            '--toggle-selected-border': 'var(--primary)',
            '--toggle-disabled-border': 'var(--button-disabled-fill)',
            '--toggle-disabled-hover-text': 'var(--disabled)',
            '--toggle-highlight-text': 'var(--common-white)',

            // Components/Checkbox
            '--checkbox-checked-active-border': 'var(--primary)',
            '--checkbox-disabled-checked-hover-fill': 'var(--checkbox-fill)',
            '--checkbox-border': 'var(--border-control)',
            '--checkbox-hover': light.secondary.main,
            '--checkbox-unchecked-fill': 'var(--surface)',
            '--checkbox-fill': common.dark,
            '--checkbox-check': common.white,
            '--checkbox-checked-fill': 'var(--checkbox-fill)',
            '--checkbox-disabled-fill': palette.neutral.light[100],
            '--checkbox-focus-ring': 'var(--secondary)',

            // Components/Input
            '--input-fill': common.white,
            '--input-disabled-fill': palette.neutral.light[100],
            '--input-disabled-invalid-border': 'var(--status-error-main)',
            '--input-outlined-hover-focus-border': 'var(--input-outlined-border-hover)',
            // '--input-label': 'var(--text-main)',
            '--input-border-default': palette.grey[700],
            '--input-border-focus': 'var(--control-accent)',
            // '--input-border-hover': 'var(--secondary)',
            '--input-border-disabled': palette.neutral.light[900],
            '--input-placeholder-text': text.placeholder,
            '--input-outlined-border-default': 'var(--border-control)',
            '--input-outlined-border-hover': 'var(--secondary)',

            // Components/Textarea
            '--textarea-disabled-fill': palette.neutral.light[100],
            '--textarea-focus-border': 'var(--textarea-border-default)',
            '--textarea-hover-focus-border': 'var(--textarea-border-hover)',
            '--textarea-fill': 'var(--surface)',
            '--textarea-border-default': 'var(--border-control)',
            '--textarea-border-hover': 'var(--secondary)',

            // Components/MultiSelect
            '--multiselect-scrollbar-thumb': 'var(--neutral-400)',

            // Components/Tooltip
            '--tooltip-fill': 'var(--neutral-light-2)',
            '--tooltip-border': palette.grey[500],

            // Components/RadioGroup
            '--radio-label-focus-fill': palette.neutral.light[200],
            '--radio-border-default': palette.neutral.light[400],
            '--radio-border-hover': 'var(--control-accent)',
            '--radio-disabled-border': palette.neutral.light[300],
            '--radio-disabled-fill': palette.neutral.light[100],
            '--radio-indicator-fill': light.primary.main,

            // Components/Input/Selectors
            // '--selector-disable-fill': common.white,
            '--select-open-fill': 'var(--primary)',
            '--select-open-text': 'var(--common-white)',
            '--select-open-placeholder-text': 'var(--common-white)',
            '--select-item-disabled-fill': palette.neutral.light[300],
            '--select-disabled-highlighted-fill': 'var(--secondary)',
            '--select-open-disabled-fill': 'var(--primary)',
            '--select-disabled-placeholder-text': 'var(--disabled)',
            '--select-invalid-placeholder-open-text': 'var(--select-open-placeholder-text)',
            '--select-open-focus-border': 'var(--primary)',
            '--select-open-invalid-hover-focus-border': 'var(--status-error-main)',
            '--select-highlight-text': 'var(--common-white)',
            '--select-placeholder-text': 'inherit',
            '--select-disabled-fill': palette.neutral.light[100],
            '--select-outlined-hover-focus-border': 'var(--input-outlined-border-hover)',
            '--select-outlined-focus-border': 'var(--input-outlined-border-default)',
            '--select-trigger-fill': 'var(--surface-subtle)',
            '--select-trigger-placeholder-text': palette.neutral.light[400],
            '--select-trigger-outlined-fill': common.white,
            '--select-border-default': palette.neutral.dark[700],
            '--select-border-focus': 'var(--control-accent)',
            '--select-content-border': palette.neutral.light[400],
            '--select-content-fill': common.white,
            '--select-item-checked-text': light.primary.main,
            '--select-separator-fill': palette.neutral.light[200],
            '--dropdown-trigger-border': palette.neutral.light[400],
            '--dropdown-popover-border': palette.neutral.light[400],
            '--dropdown-popover-fill': 'var(--surface)',
            '--dropdown-option-hover-fill': 'var(--surface-hover)',
            '--dropdown-option-disabled-fill': palette.neutral.light[300],
            '--dropdown-tooltip-fill': 'var(--surface-hover)',

            // Components/Data Display/ DataTable
            '--data-table-fill': 'var(--surface-subtle)',
            '--data-table-header-fill': 'var(--surface-subtle)',
            '--data-table-row-even-fill': palette.neutral.light[200],
            '--data-table-row-odd-fill': 'var(--surface-subtle)',
            '--data-table-row-hover-fill': 'var(--surface-hover)',
            '--data-table-row-selected-outline': light.primary.main,
            '--data-table-row-selected': 'inset 0px 0px 0px 2px var(--data-table-row-selected-outline)',
            '--data-table-wrapper-border': 'var(--neutral-light-4)',
            '--data-table-resize-fill': 'var(--neutral-5)',
            '--data-table-resize-active-fill': 'var(--primary)',

            // Components/Switch
            '--switch-fill': palette.neutral.dark[700],
            '--switch-disabled-fill': palette.neutral.light[300],
            '--switch-thumb-fill': palette.neutral.light[50],

            // Components/Button
            '--button-secondary-fill': palette.neutral.light[300],
            '--button-secondary-active-fill': palette.neutral.light[400],
            '--button-icon-fill': palette.neutral.light[400],
            '--button-disabled-fill': palette.neutral.light[200],
            '--button-transparent-border': palette.neutral.light[400],
            '--button-disabled-text': palette.grey[700],
            '--button-primary-text': 'var(--common-white)',
            '--button-secondary-text': 'var(--common-dark)',
            '--button-highlight-text': 'var(--common-white)',
            '--button-primary-active-fill': light.primary.variant,
            '--button-text-active': light.primary.variant,

            // Components/Menu
            '--menu-content-fill': common.white,
            '--menu-item-highlight-fill': light.secondary.main,
            '--menu-item-active-fill': light.primary.main,
            '--menu-item-disabled-fill': palette.neutral.light[200],
            '--menu-separator-fill': palette.neutral.light[300],
            '--menu-content-border': palette.neutral.light[300],
            '--menu-sub-content-border': palette.neutral.light[200],
            '--menu-item-highlight-border': light.secondary.main,
            '--menu-item-highlight-text': common.white,
            '--menu-item-disabled-text': palette.grey[700],

            // Components/Data Display/Badge & Chip / Badge
            '--badge-primary-surface': light.badge.primary.fill,
            '--badge-primary-accent': light.badge.primary.outline,
            '--badge-secondary-surface': light.badge.secondary.fill,
            '--badge-secondary-accent': light.badge.secondary.outline,
            '--badge-grey-surface': light.badge.grey.fill,
            '--badge-grey-accent': light.badge.grey.outline,
            '--badge-red-surface': light.badge.red.fill,
            '--badge-red-accent': light.badge.red.outline,
            '--badge-orange-surface': light.badge.orange.fill,
            '--badge-orange-accent': light.badge.orange.outline,
            '--badge-green-surface': light.badge.green.fill,
            '--badge-green-accent': light.badge.green.outline,
            '--badge-blue-surface': light.badge.blue.fill,
            '--badge-blue-accent': light.badge.blue.outline,
            '--badge-error-surface': 'var(--status-error-main)',
            '--badge-warning-surface': 'var(--status-warning-main)',
            '--badge-success-surface': 'var(--status-success-main)',

            // Components/Data Display/Badge & Chip / Chip
            // '--chip-indeterminate': palette.neutral.light[300],
            // '--chip-indeterminate-hover': palette.neutral.light[600],
            // '--chip-outline-fill': common.white,
            // '--chip-outline-hover': palette.neutral.light[100],
            // '--chip-outline-active': palette.neutral.light[300],

            // Components/ SideNav
            // '--sidenav-bg': palette.neutral.light[50],
            // '--sidenav-bg-hover': palette.neutral.light[200],
            // '--sidenav-bg-active': palette.neutral.light[300],
            // '--nav-control-btn': palette.neutral.light[300],
            // '--nav-control-btn-hover': palette.neutral.light[600],
            // '--nav-control-btn-focus': palette.neutral.light[500],

            // Components/ SideNav/ Nav items
            // '--nav-item-default': 'var(--text-main)',
            // '--nav-item-hover': 'var(--secondary)',
            // '--nav-item-active': 'var(--primary)',

            // Components / Graph
            // '--platform-badge-bg': elevation.light[2],

            // Components / Graph / Edge
            // '--edge-color': light.edge.color,
            // '--edge-label-bg': palette.neutral.light[200],
            // '--edge-label-text': 'var(--text-main)',

            // Utilities
            // '--icon': 'var(--text-main)',
            // '--icon-contrast': 'var(--text-contrast)',
            '--icon-disabled': palette.grey[700],
            // '--divider': palette.neutral.light[500],

            // Legacy below
            // same as palette.neutral.dark[50]
            '--contrast': '#121212',

            '--secondary-variant-2': secondaryVariant2,
            // '--tertiary': '#02c577',
            // '--tertiary-variant': '#5cc791',

            '--neutral-1': 'var(--neutral-light-1)',
            '--neutral-2': 'var(--neutral-light-2)',
            '--neutral-3': 'var(--neutral-light-3)',
            '--neutral-4': 'var(--neutral-light-4)',
            '--neutral-5': 'var(--neutral-light-5)',

            '--link': '#1a30ff',

            '--error': '#b44641',

            '--neutral-light-1': common.white,
            '--neutral-light-2': palette.neutral.light[100],
            '--neutral-light-3': palette.neutral.light[200],
            '--neutral-light-4': palette.neutral.light[300],
            '--neutral-light-5': palette.neutral.light[400],

            '--neutral-dark-1': palette.neutral.dark[50],
            '--neutral-dark-2': palette.neutral.dark[400],
            '--neutral-dark-3': palette.neutral.dark[500],
            '--neutral-dark-4': palette.neutral.dark[600],
            '--neutral-dark-5': palette.neutral.dark[700],
        },

        '.dark': {
            // MAIN
            '--primary': dark.primary.main,
            '--primary-variant': dark.primary.variant,
            '--secondary': dark.secondary.main,
            '--secondary-variant': dark.secondary.variant,
            '--tertiary': dark.tertiary.main,
            '--tertiary-variant': dark.tertiary.variant,
            '--disabled': dark.disabled,

            // Shared control and surface tokens
            '--border-control': dark.input.border,
            '--control-accent': secondaryVariant2,
            '--surface': common.dark,
            '--surface-subtle': palette.neutral.dark[400],
            '--surface-hover': palette.neutral.dark[600],

            // NEUTRALS
            '--neutral-50': palette.neutral.dark[50],
            '--neutral-100': palette.neutral.dark[100],
            '--neutral-200': palette.neutral.dark[200],
            '--neutral-300': palette.neutral.dark[300],
            '--neutral-400': palette.neutral.dark[400],
            '--neutral-500': palette.neutral.dark[500],
            '--neutral-600': palette.neutral.dark[600],
            '--neutral-700': palette.neutral.dark[700],
            '--neutral-800': palette.neutral.dark[800],
            '--neutral-900': palette.neutral.dark[900],

            // TEXT
            '--text-main': common.white,
            '--text-light': text.dark,
            '--text-contrast': common.dark,
            '--text-faint': dark.text.faint,

            // '--text-primary': dark.primary.main,
            // '--text-secondary': dark.secondary.main,

            // LINKS
            '--link-main': 'var(--secondary)',
            '--link-hover': 'var(--secondary-variant)',

            // FOCUS
            '--focus-ring': 'var(--secondary)',
            '--focus-ring-offset': palette.neutral.dark[50],

            // ELEVATION
            // '--elevation-0': elevation.dark[0],
            '--elevation-1': elevation.dark[1],
            '--elevation-2': elevation.dark[2],
            '--elevation-3': elevation.dark[3],
            '--elevation-4': elevation.dark[4],
            '--elevation-5': elevation.dark[5],

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

            // BRAND COLORS
            '--bhe-main': 'var(--primary)',
            // '--sp-main': 'var(--primary)',
            '--bhce-main': dark['bhce-main'],
            // '--logo-neutral': common.white,
            // '--brand-primary-dark-purple': 'var(--primary-variant)',
            // '--brand-primary-deep-purple': 'var(--primary)',
            // '--brand-primary-medium-blue': 'var(--secondary)',
            // '--brand-secondary-highlight-green': 'var(--tertiary)',

            // Components/Accordion
            '--accordion-fill': 'var(--neutral-dark-2)',

            // Components/Card
            '--card-border': 'var(--neutral-900)',

            // Components/Table
            '--table-fill': 'var(--neutral-dark-2)',
            '--table-row-border-width': '0px',
            '--table-dragging-fill': 'var(--neutral-dark-1)',

            // Components/Tabs
            '--tabs-text': common.white,
            '--tabs-active-accent': 'var(--secondary-variant-2)',

            // Components/Typography
            '--typography-body-text': 'var(--text-main)',

            // Components/Breadcrumbs
            '--breadcrumb-text': common.white,
            '--breadcrumb-page-text': 'var(--secondary-variant-2)',

            // Components/DatePicker
            '--date-picker-input-fill': 'var(--common-dark)',
            '--date-picker-content-border': 'var(--neutral-light-1)',

            // Components/Pagination
            '--pagination-text': common.white,
            '--pagination-disabled-text': palette.neutral.light[800],

            // Components/Skeleton
            '--skeleton-fill': 'var(--neutral-dark-5)',

            // Components/ComboBadge
            '--combo-badge-fill': 'var(--neutral-dark-5)',
            '--combo-badge-text': 'var(--neutral-light-1)',

            // Components/Dialog
            '--dialog-overlay-fill': 'rgb(0 0 0 / 0.4)',
            '--dialog-fill': 'var(--neutral-dark-2)',
            '--dialog-text': 'var(--neutral-light-1)',

            // Components/Popover
            '--popover-fill': 'var(--neutral-dark-2)',
            '--popover-text': common.white,

            // Components/Slider
            '--slider-track-fill': 'var(--neutral-200)',
            '--slider-thumb-fill': 'var(--neutral-100)',
            '--slider-thumb-border': 'var(--neutral-300)',
            '--slider-thumb-focus-fill': 'var(--neutral-400)',
            '--slider-thumb-active-fill': 'var(--common-white)',

            // Components/Stepper
            '--stepper-inactive-border': 'var(--input-border-default)',

            // Components/Toggle
            '--toggle-btn-fill': 'var(--surface)',
            '--toggle-btn-border': palette.neutral.dark[600],
            '--toggle-group-fill': 'var(--surface-subtle)',

            '--toggle-hover-border': 'var(--toggle-btn-border)',
            '--toggle-active-border': 'var(--toggle-btn-border)',
            '--toggle-selected-border': 'var(--toggle-btn-border)',
            '--toggle-disabled-border': 'var(--toggle-btn-border)',
            '--toggle-disabled-hover-text': 'var(--toggle-highlight-text)',
            '--toggle-highlight-text': 'var(--common-dark)',

            // Components/Checkbox
            '--checkbox-checked-active-border': 'var(--primary-variant)',
            '--checkbox-disabled-checked-hover-fill': 'var(--checkbox-hover)',
            '--checkbox-border': 'var(--border-control)',
            '--checkbox-hover': dark.badge.secondary.outline,
            '--checkbox-unchecked-fill': 'var(--surface)',
            '--checkbox-fill': common.white,
            '--checkbox-check': common.dark,
            '--checkbox-checked-fill': 'var(--primary)',
            '--checkbox-disabled-fill': palette.neutral.dark[400],
            '--checkbox-focus-ring': secondaryVariant2,

            // Components/Input
            '--input-fill': palette.neutral.dark[50],
            '--input-disabled-fill': palette.neutral.dark[400],
            '--input-disabled-invalid-border': 'var(--input-border-disabled)',
            '--input-outlined-hover-focus-border': 'var(--input-outlined-border-default)',
            // '--input-label': 'var(--text-main)',
            '--input-border-default': palette.neutral.light[400],
            '--input-border-focus': 'var(--control-accent)',
            // '--input-border-hover': 'var(--secondary)',
            '--input-border-disabled': palette.neutral.dark[900],

            '--input-placeholder-text': dark.input.placeholder,
            '--input-outlined-border-default': 'var(--border-control)',
            '--input-outlined-border-hover': 'var(--secondary)',

            // Components/Textarea
            '--textarea-disabled-fill': palette.neutral.dark[400],
            '--textarea-focus-border': 'var(--textarea-border-default)',
            '--textarea-hover-focus-border': 'var(--textarea-border-default)',
            '--textarea-fill': 'var(--surface)',
            '--textarea-border-default': 'var(--border-control)',
            '--textarea-border-hover': 'var(--secondary)',

            // Components/MultiSelect
            '--multiselect-scrollbar-thumb': 'var(--input-placeholder-text)',

            // Components/Tooltip
            '--tooltip-fill': 'var(--neutral-dark-5)',
            '--tooltip-border': palette.grey[500],

            // Components/RadioGroup
            '--radio-label-focus-fill': palette.neutral.dark[600],
            '--radio-border-default': palette.neutral.light[400],
            '--radio-border-hover': 'var(--control-accent)',
            '--radio-disabled-border': palette.neutral.dark[600],
            '--radio-disabled-fill': palette.neutral.dark[400],
            '--radio-indicator-fill': secondaryVariant2,

            // Components/Input/Selectors
            // '--selector-disable-fill': common.white,
            '--select-open-fill': 'var(--input-fill)',
            '--select-open-text': 'var(--text-main)',
            '--select-open-placeholder-text': 'var(--input-placeholder-text)',
            '--select-item-disabled-fill': palette.neutral.dark[700],
            '--select-disabled-highlighted-fill': 'var(--select-item-disabled-fill)',
            '--select-open-disabled-fill': 'var(--select-disabled-fill)',
            '--select-disabled-placeholder-text': 'var(--select-trigger-placeholder-text)',
            '--select-invalid-placeholder-open-text': 'var(--select-open-placeholder-text)',
            '--select-open-focus-border': 'var(--input-outlined-border-default)',
            '--select-open-invalid-hover-focus-border': 'var(--status-error-main)',
            '--select-highlight-text': 'var(--common-dark)',
            '--select-placeholder-text': 'var(--select-trigger-placeholder-text)',
            '--select-disabled-fill': palette.neutral.dark[400],
            '--select-outlined-hover-focus-border': 'var(--input-outlined-border-default)',
            '--select-outlined-focus-border': 'var(--input-outlined-border-default)',
            '--select-trigger-fill': 'var(--surface-subtle)',
            '--select-trigger-placeholder-text': dark.input.placeholder,
            '--select-trigger-outlined-fill': palette.neutral.dark[50],
            '--select-border-default': palette.neutral.light[400],
            '--select-border-focus': 'var(--control-accent)',
            '--select-content-border': palette.neutral.light[900],
            '--select-content-fill': palette.neutral.dark[400],
            '--select-item-checked-text': secondaryVariant2,
            '--select-separator-fill': palette.neutral.light[200],
            '--dropdown-trigger-border': palette.neutral.light[400],
            '--dropdown-popover-border': palette.neutral.dark[700],
            '--dropdown-popover-fill': 'var(--surface)',
            '--dropdown-option-hover-fill': 'var(--surface-hover)',
            '--dropdown-option-disabled-fill': palette.neutral.dark[700],
            '--dropdown-tooltip-fill': 'var(--surface-hover)',

            // Components/Data Display/ DataTable
            '--data-table-fill': 'var(--surface-subtle)',
            '--data-table-header-fill': 'var(--surface-subtle)',
            '--data-table-row-even-fill': palette.neutral.dark[500],
            '--data-table-row-odd-fill': 'var(--surface-subtle)',
            '--data-table-row-hover-fill': 'var(--surface-hover)',
            '--data-table-row-selected-outline': darkDataTableRowSelectedOutline,
            '--data-table-row-selected': 'inset 0px 0px 0px 2px var(--data-table-row-selected-outline)',
            '--data-table-wrapper-border': 'var(--neutral-900)',
            '--data-table-resize-fill': 'var(--neutral-light-5)',
            '--data-table-resize-active-fill': 'var(--secondary)',

            // Components/Switch
            '--switch-fill': common.white,
            '--switch-disabled-fill': common.disabled,
            '--switch-thumb-fill': palette.neutral.dark[50],

            // Components/Button
            '--button-secondary-fill': palette.neutral.dark[700],
            '--button-secondary-active-fill': palette.neutral.dark[700],
            '--button-icon-fill': palette.neutral.light[400],
            '--button-disabled-fill': palette.neutral.dark[700],
            '--button-transparent-border': palette.neutral.dark[700],
            '--button-disabled-text': common.disabled,
            '--button-primary-text': 'var(--common-dark)',
            '--button-secondary-text': 'var(--common-white)',
            '--button-highlight-text': 'var(--common-dark)',
            '--button-primary-active-fill': dark.primary.variant,
            '--button-text-active': 'var(--primary)',

            // Components/Menu
            '--menu-content-fill': palette.neutral.dark[400],
            '--menu-item-highlight-fill': dark.secondary.main,
            '--menu-item-active-fill': dark.primary.main,
            '--menu-item-disabled-fill': palette.neutral.dark[700],
            '--menu-separator-fill': palette.neutral.dark[600],
            '--menu-content-border': palette.neutral.dark[600],
            '--menu-sub-content-border': palette.neutral.dark[600],
            '--menu-item-highlight-border': dark.secondary.main,
            '--menu-item-highlight-text': common.black,
            '--menu-item-disabled-text': common.disabled,

            // Components/Data Display/Badge & Chip / Badge
            '--badge-primary-surface': dark.badge.primary.fill,
            '--badge-primary-accent': dark.badge.primary.outline,
            '--badge-secondary-surface': dark.badge.secondary.fill,
            '--badge-secondary-accent': dark.badge.secondary.outline,
            '--badge-grey-surface': dark.badge.grey.fill,
            '--badge-grey-accent': dark.badge.grey.outline,
            '--badge-red-surface': dark.badge.red.fill,
            '--badge-red-accent': dark.badge.red.outline,
            '--badge-orange-surface': dark.badge.orange.fill,
            '--badge-orange-accent': dark.badge.orange.outline,
            '--badge-green-surface': dark.badge.green.fill,
            '--badge-green-accent': dark.badge.green.outline,
            '--badge-blue-surface': dark.badge.blue.fill,
            '--badge-blue-accent': dark.badge.blue.outline,
            '--badge-error-surface': 'var(--status-error-fill)',
            '--badge-warning-surface': 'var(--status-warning-fill)',
            '--badge-success-surface': 'var(--status-success-fill)',

            // Components/Data Display/Badge & Chip / Chip
            // '--chip-indeterminate': palette.neutral.dark[500],
            // '--chip-indeterminate-hover': palette.neutral.dark[300],
            // '--chip-outline-fill': elevation.dark[1],
            // '--chip-outline-hover': palette.neutral.dark[200],
            // '--chip-outline-active': palette.neutral.dark[400],

            // Components/ SideNav
            // '--sidenav-bg': palette.neutral.dark[300],
            // '--sidenav-bg-hover': palette.neutral.dark[200],
            // '--sidenav-bg-active': palette.neutral.dark[600],
            // '--nav-control-btn': palette.neutral.dark[300],
            // '--nav-control-btn-hover': palette.neutral.dark[500],
            // '--nav-control-btn-focus': palette.neutral.dark[400],

            // Components/ SideNav/ Nav items
            // '--nav-item-default': 'var(--text-main)',
            // '--nav-item-hover': 'var(--secondary)',
            // '--nav-item-active': 'var(--primary)',

            // Components / Graph
            // '--platform-badge-bg': elevation.dark[2],

            // Components / Graph / Edge
            // '--edge-color': dark.edge.color,
            // '--edge-label-bg': palette.neutral.dark[500],
            // '--edge-label-text': 'var(--text-main)',

            // Utilities
            // '--icon': 'var(--text-main)',
            // '--icon-contrast': 'var(--text-contrast)',
            '--icon-disabled': common.disabled,
            // '--divider': palette.neutral.dark[500],

            ///////////// Legacy below //////////
            // same as common.white
            '--contrast': common.white,
            '--neutral-1': 'var(--neutral-dark-1)',
            '--neutral-2': 'var(--neutral-dark-2)',
            '--neutral-3': 'var(--neutral-dark-3)',
            '--neutral-4': 'var(--neutral-dark-4)',
            '--neutral-5': 'var(--neutral-dark-5)',

            '--link': secondaryVariant2,

            '--error': '#e9827c',
        },
    });

    addUtilities({
        '.focus-ring': {
            outline: 'var(--focus-ring-width) solid var(--focus-ring)',
            'outline-offset': 'var(--focus-ring-offset-width)',
            'box-shadow': 'var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow, 0 0 #0000)',
        },
        '.focus-ring-inset': {
            outline: 'none',
            'box-shadow': 'inset 0 0 0 var(--focus-ring-width) var(--focus-ring), var(--tw-shadow, 0 0 #0000)',
        },
        '.clip-right-rounded': {
            'clip-path': 'inset(0 0.5px 0 -100vw round 0.25rem)',
        },
        '.clip-left-rounded': {
            'clip-path': 'inset(0 -100vw 0 0 round 0.25rem)',
        },
    });
};

export default plugin;
