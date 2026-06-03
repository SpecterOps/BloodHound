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

export type Theme = 'light' | 'dark';

export const THEMES: readonly Theme[] = ['light', 'dark'] as const;

// Worker-scoped Playwright option shape consumed by `defineConfig<TestOptions>` callers.
// The fixture that declares the `theme` option lives in `./axe` so consumers that import
// the shared `test` automatically get the option without needing a second composition step.
export type TestOptions = {
    theme: Theme;
};

// Canonical path (relative to the consumer's Playwright project root) for a per-theme
// `storageState` snapshot. `loginAndSnapshotThemes` in `./auth` writes to these paths and
// project configs read from them via `use.storageState`.
export const authStorageStateFor = (theme: Theme): string => `./playwright/.auth/user-${theme}.json`;
