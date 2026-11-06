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

export const tableClasses = 'w-full caption-bottom text-sm';

export const tableHeaderClasses = '[&_tr]:border-b bg-[color:var(--table-fill)]';

export const tableBodyClasses = '[&_tr:last-child]:border-0 bg-[color:var(--table-fill)]';

export const tableFooterClasses = 'border-t bg-muted/50 font-medium [&>tr]:last:border-b-0';

export const tableRowClasses =
    'border-b-[length:var(--table-row-border-width)] transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted';

export const tableHeadClasses =
    'h-12 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0';

export const tableCellClasses = 'p-4 pl-0 align-middle [&:has([role=checkbox])]:pr-0';

export const tableCaptionClasses = 'mt-4 text-sm text-muted-foreground';

export const tableDraggingCellClasses = '!bg-[color:var(--table-dragging-fill)]';
