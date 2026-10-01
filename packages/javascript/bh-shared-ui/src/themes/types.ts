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

export type ThemeMode = 'light' | 'dark';

export type ThemeColors = {
    primary: string;
    primaryVariant: string;
    secondary: string;
    secondaryVariant: string;
    onPrimary: string;
    onSecondary: string;
    text: string;
    mutedText: string;
    contrast: string;
    link: string;
    neutral: [string, string, string, string, string];
};

export type ThemePackage = {
    schemaVersion: 1;
    id: string;
    version: string;
    name: string;
    author: string;
    description: string;
    variants: Record<ThemeMode, ThemeColors>;
    availableEffects?: Array<'clippy' | 'retro-cursor'>;
};

export type ThemeEffect = NonNullable<ThemePackage['availableEffects']>[number];

export type ThemeCatalogEntry = {
    id: string;
    name: string;
    description: string;
    author: string;
    packagePath: string;
    preview: [string, string, string];
};
