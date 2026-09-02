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
export const themeOptions = [
    'primary',
    'primary-variant',
    'primary',
    'primary-variant',
    'secondary',
    'secondary-variant',
    'tertiary',
    'tertiary-variant',
] as const;

export type ThemeOptions = (typeof themeOptions)[number];

// this means we cant use predefined/cross-browser color names. But values from figma work fine
export type CustomColorOptions =
    | `#${string}`
    | `rgb(${string})`
    | `rgba(${string})`
    | `hsl(${string})`
    | `hsla(${string})`;

export type ColorOptions = ThemeOptions | CustomColorOptions;
