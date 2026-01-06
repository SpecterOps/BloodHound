// Copyright 2025 Specter Ops, Inc.
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
export type ShortCutsMap = Record<string, string[][]>;

export const EXPLORE_SHORTCUTS = {
    'Explore Page': [
        ['/', 'Jump to Node Search'],
        ['P', 'Jump to Pathfinding'],
        ['C', 'Focus Cypher Query Editor'],
        ['S', 'Save Current Query'],
        ['R', 'Run Current Cypher Query'],
        ['Shift + /', 'Search Current Nodes'],
        ['T', 'Toggle Table View'],
        ['I', 'Toggle Node Info Panel'],
        ['G', 'Reset Graph View'],
    ],
};

export const GLOBAL_SHORTCUTS = {
    Global: [
        ['[1-5]', 'Navigate sidebar pages'],
        ['D', 'Navigate to Documentation'],
        ['H', 'Launch keyboard shortcuts list'],
        ['U', 'Launch File Upload dialog'],
        ['M', 'Toggle Dark Mode'],
    ],
};

export const POSTURE_PAGE_SHORTCUTS = {
    'Posture Page': [['F', 'Filter Table Data']],
};
