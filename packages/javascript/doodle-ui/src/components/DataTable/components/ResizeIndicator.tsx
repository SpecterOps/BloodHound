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
import { useSortable } from '@dnd-kit/sortable';
import { Header } from '@tanstack/table-core';
import { cn } from '../../utils';

interface ResizeIndicatorProps<TData, TValue> {
    header: Header<TData, TValue>;
    tableHeight: number | null;
    onSetExpanded: (header: Header<TData, TValue>) => void;
}

export const ResizeIndicator = <TData, TValue>({
    header,
    tableHeight,
    onSetExpanded,
}: ResizeIndicatorProps<TData, TValue>) => {
    const { isDragging } = useSortable({ id: header.column.id });

    if (isDragging) return null;
    return (
        <div
            {...{
                onDoubleClick: () => onSetExpanded(header),
                onMouseDown: header.getResizeHandler(),
                onTouchStart: header.getResizeHandler(),
                className: cn(
                    `resizer opacity-0 absolute top-0 right-0 w-1 bg-neutral-5 dark:bg-neutral-light-5 cursor-col-resize select-none touch-none group-hover:opacity-100 group-[.isResizing~_&:hover]:opacity-0 group-[.isResizing~_&:hover]:cursor-default [th:has(~_.isResizing)_>_&]:!opacity-0 [th:has(~_.isResizing)_>_&]:cursor-default`,
                    header.column.getIsResizing() ? 'bg-primary opacity-100 dark:bg-secondary' : ''
                ),
            }}
            style={{
                height: tableHeight ? `${tableHeight}px` : '100%',
                transform: 'translateZ(0)',
            }}
        />
    );
};
