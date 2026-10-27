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

/**
 * Resolves a doc-relative (or repo-root-relative, when it starts with "/") path to a repo path.
 * Returns null when the path climbs above the repo root or contains backslashes.
 */
export const resolveRepoPath = (currentDocPath: string, target: string) => {
    const segments = target.startsWith('/') ? [] : currentDocPath.split('/').slice(0, -1);
    for (const segment of safeDecode(target).split('/')) {
        if (segment.includes('\\')) return null;
        if (segment === '' || segment === '.') continue;
        if (segment === '..') {
            if (segments.length === 0) return null;
            segments.pop();
        } else {
            segments.push(segment);
        }
    }
    return segments.join('/');
};

/**
 * Decodes a URI component, returning the original value when it is not valid percent-encoding.
 */
const safeDecode = (value: string) => {
    try {
        return decodeURIComponent(value);
    } catch {
        return value;
    }
};
