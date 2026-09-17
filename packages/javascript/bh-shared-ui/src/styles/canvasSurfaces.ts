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

/** Matches a page-level surface to the application canvas in both themes. */
export const CANVAS_SURFACE_CLASS = '!bg-neutral-1 !shadow-none';

/** Adds the standard neutral boundary used by page-level content surfaces. */
export const OUTLINED_CANVAS_SURFACE_CLASS = `${CANVAS_SURFACE_CLASS} border border-solid border-neutral-light-4 dark:border-neutral-900`;

/** Keeps table rows on the canvas while preserving stateful row styling. */
export const CANVAS_TABLE_ROW_CLASS =
    'bg-neutral-1 hover:bg-neutral-3 data-[state=selected]:bg-muted aria-selected:bg-muted';

/** Applies the canvas treatment to legacy MUI table containers and their non-selected rows. */
export const CANVAS_MUI_TABLE_CLASS = `${OUTLINED_CANVAS_SURFACE_CLASS} [&_.MuiTableBody-root]:!bg-neutral-1 [&_.MuiTableRow-root:not(.Mui-selected)]:!bg-neutral-1 [&_.MuiTableRow-hover:not(.Mui-selected):hover]:!bg-neutral-3`;
