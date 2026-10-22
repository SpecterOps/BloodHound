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
import { palette } from '../../tailwind/colors';

export const dataTableBackgroundColors = {
    'data-table-fill': 'var(--surface-subtle)',
    'data-table-header-fill': 'var(--surface-subtle)',
    'data-table-row-even-fill': palette.neutral.light[200],
    'data-table-row-even-fill-dark': palette.neutral.dark[500],
    'data-table-row-odd-fill': 'var(--surface-subtle)',
    'data-table-row-hover-fill': 'var(--surface-hover)',
};

export const dataTableBoxShadows = {
    'data-table-row-selected': 'inset 0px 0px 0px 2px var(--primary)',
    'data-table-row-selected-dark': 'inset 0px 0px 0px 2px #4A42B5',
};

export const dataTableWrapperClasses =
    'w-full border border-solid border-neutral-light-4 bg-neutral-1 shadow-none dark:border-neutral-900';
export const dataTableClasses = 'after:block after:h-[var(--prevent-vanishing-sticky-header)]';
export const dataTableHeaderClasses = '[&_tr]:border-0';
export const dataTableHeaderRowClasses = 'border-0';
export const dataTableHeadClasses = 'bg-neutral-2 text-nowrap group relative z-10 overflow-x-clip';
export const dataTableRowClasses = 'bg-neutral-1 hover:bg-data-table-row-hover-fill';
export const dataTableSelectedRowClasses = 'shadow-data-table-row-selected dark:shadow-data-table-row-selected-dark';
export const dataTableClickableRowClasses =
    'focus:outline-none focus-visible:focus-ring focus-visible:bg-data-table-row-hover-fill';
export const dataTableCellClasses = 'text-left overflow-x-clip -outline-offset-2';
