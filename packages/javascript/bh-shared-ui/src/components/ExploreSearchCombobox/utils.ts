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

import { SearchResult } from '../../hooks';

// Returns the set of display names (name, falling back to objectid) that appear on more than one
// search result so the distinguished name should be shown on those results to tell them apart
export const getDuplicateDisplayNames = (results: SearchResult[]): Set<string> => {
    const displayNameCounts = new Map<string, number>();

    for (const result of results) {
        const displayName = result.name || result.objectid;

        if (displayName) {
            displayNameCounts.set(displayName, (displayNameCounts.get(displayName) ?? 0) + 1);
        }
    }

    return new Set(
        [...displayNameCounts.entries()].filter(([, count]) => count > 1).map(([displayName]) => displayName)
    );
};
