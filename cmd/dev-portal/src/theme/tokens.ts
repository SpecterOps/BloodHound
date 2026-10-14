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

// The names every theme must define. Tailwind exposes each one as a `dp-` utility (e.g. `bg-dp-canvas`,
// `font-dp-display`, `rounded-dp-sm`) backed by a CSS custom property, so components never hold raw values.
// This file is also loaded by tailwind.config.js, so it must stay free of browser APIs.

export const colorTokens = [
    // Page chrome and reading surfaces
    'canvas',
    'surface',
    'sunken',
    'border',
    'border-strong',
    'text',
    'text-muted',
    'link',
    'link-hover',
    'accent',
    'focus',
    // "You are here": the graph's source node and the active Doc Page
    'marker',
    // The landing page's graph canvas
    'hero-canvas',
    'hero-text',
    'hero-text-muted',
    'hero-edge',
    'hero-node',
] as const;

export const fontTokens = ['display', 'body', 'code'] as const;

export const radiusTokens = ['sm', 'md'] as const;

export type ColorToken = (typeof colorTokens)[number];
export type FontToken = (typeof fontTokens)[number];
export type RadiusToken = (typeof radiusTokens)[number];

export type Theme = {
    // Tells the browser and DoodleUI components whether to use their light or dark styles
    scheme: 'light' | 'dark';
    colors: Record<ColorToken, string>;
    fonts: Record<FontToken, string>;
    radii: Record<RadiusToken, string>;
};

export const colorVar = (token: ColorToken) => `--dp-color-${token}`;
export const fontVar = (token: FontToken) => `--dp-font-${token}`;
export const radiusVar = (token: RadiusToken) => `--dp-radius-${token}`;

// Maps each token to `var(--dp-…)`, keyed by token name, for the Tailwind theme
const toVarMap = <Token extends string>(tokens: readonly Token[], toVar: (token: Token) => string) =>
    Object.fromEntries(tokens.map((token) => [token, `var(${toVar(token)})`])) as Record<Token, string>;

export const tailwindTokens = {
    colors: toVarMap(colorTokens, colorVar),
    fonts: toVarMap(fontTokens, fontVar),
    radii: toVarMap(radiusTokens, radiusVar),
};

export const themeCustomProperties = (theme: Theme): Record<string, string> => ({
    ...Object.fromEntries(colorTokens.map((token) => [colorVar(token), theme.colors[token]])),
    ...Object.fromEntries(fontTokens.map((token) => [fontVar(token), theme.fonts[token]])),
    ...Object.fromEntries(radiusTokens.map((token) => [radiusVar(token), theme.radii[token]])),
});
