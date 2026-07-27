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

import { useLocalStorage } from './useLocalStorage';

/**
 * The localStorage key backing the main navigation's expanded/collapsed state.
 * Owned here so both the nav and any consumers that need to react to the nav's
 * width share a single source of truth instead of duplicating the literal.
 */
export const NAV_EXPANDED_STORAGE_KEY = 'isNavExpanded';

/** Default expanded state used when the key is absent from localStorage. */
export const NAV_EXPANDED_DEFAULT = true;

/**
 * Syncs with the main navigation's expanded/collapsed state persisted in
 * localStorage. Returns a `[value, setValue]` tuple like `useState`; consumers
 * that only need to read the state can ignore the setter.
 */
export const useNavExpanded = () => useLocalStorage<boolean>(NAV_EXPANDED_STORAGE_KEY, NAV_EXPANDED_DEFAULT);
