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

import { themeCustomProperties, type Theme } from './tokens';

// Typefaces are loaded in main.tsx; a theme that names another family must load it there too
const fonts: Theme['fonts'] = {
    display: '"Nunito Sans", "Avenir Next", "Segoe UI", Helvetica, Arial, sans-serif',
    body: 'Figtree, "Segoe UI", Helvetica, Arial, sans-serif',
    code: 'ui-monospace, "SF Mono", Menlo, Consolas, monospace',
};

const radii: Theme['radii'] = {
    sm: '4px',
    md: '8px',
};

// SpecterOps brand palette
const brand = {
    inkIndigo: '#0D0A30',
    deepPurple: '#2C2677',
    mediumBlue: '#4A3BD7',
    neonBlue: '#5465FF',
    lavender: '#D7D6E6',
    signalOrange: '#F46036',
};

const specter: Theme = {
    scheme: 'light',
    colors: {
        canvas: '#F7F7FB',
        surface: '#FFFFFF',
        sunken: '#EEEDF6',
        border: brand.lavender,
        'border-strong': '#B4B2CF',
        text: brand.inkIndigo,
        'text-muted': '#55537A',
        link: brand.mediumBlue,
        'link-hover': brand.deepPurple,
        accent: brand.deepPurple,
        focus: brand.mediumBlue,
        marker: brand.signalOrange,
        'hero-canvas': brand.inkIndigo,
        'hero-text': '#FFFFFF',
        'hero-text-muted': '#B9B7DA',
        'hero-edge': brand.neonBlue,
        'hero-node': '#1B1650',
    },
    fonts,
    radii,
};

const specterNight: Theme = {
    scheme: 'dark',
    colors: {
        canvas: brand.inkIndigo,
        surface: '#141040',
        sunken: '#1C1752',
        border: '#2E2966',
        'border-strong': '#463F8C',
        text: '#ECEBF7',
        'text-muted': '#A9A7CC',
        link: '#8F98FF',
        'link-hover': '#B8BEFF',
        accent: '#8F98FF',
        focus: '#8F98FF',
        marker: brand.signalOrange,
        'hero-canvas': '#07051F',
        'hero-text': '#FFFFFF',
        'hero-text-muted': '#B9B7DA',
        'hero-edge': brand.neonBlue,
        'hero-node': '#15114A',
    },
    fonts,
    radii,
};

export const themes = { specter, specterNight } satisfies Record<string, Theme>;

export type ThemeName = keyof typeof themes;

// The theme shown for each operating system color scheme preference
export const themeForScheme: Record<Theme['scheme'], ThemeName> = {
    light: 'specter',
    dark: 'specterNight',
};

export const applyTheme = (themeName: ThemeName, root: HTMLElement = document.documentElement) => {
    const theme = themes[themeName];
    Object.entries(themeCustomProperties(theme)).forEach(([property, value]) =>
        root.style.setProperty(property, value)
    );
    root.dataset.theme = themeName;
    root.style.colorScheme = theme.scheme;
    root.classList.toggle('dark', theme.scheme === 'dark');
};

// Applies the theme matching the operating system preference and follows later changes to it
export const followSystemTheme = () => {
    const darkQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const applyForQuery = () => applyTheme(themeForScheme[darkQuery.matches ? 'dark' : 'light']);
    applyForQuery();
    darkQuery.addEventListener('change', applyForQuery);
};
