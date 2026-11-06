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

import { type RefObject, useLayoutEffect } from 'react';

export const useTableLayout = (
    viewportRef: RefObject<HTMLDivElement>,
    headerRef: RefObject<HTMLTableSectionElement>,
    bodyRef: RefObject<HTMLTableSectionElement>,
    contentHeight: number
) => {
    useLayoutEffect(() => {
        const viewport = viewportRef.current;
        const header = headerRef.current;
        const body = bodyRef.current;

        if (!viewport || !header || !body) return;

        const updateHeight = () => {
            // Measure the viewport without the extra height from the previous layout.
            // Otherwise auto-height tables cannot shrink, and padding causes a resize loop.
            viewport.style.setProperty('--data-table-content-height', `${contentHeight}px`);
            viewport.style.setProperty('--prevent-vanishing-sticky-header', '0px');

            const headerHeight = header.clientHeight;
            const bodyHeight = body.getBoundingClientRect().height;
            const viewportStyle = getComputedStyle(viewport);
            const viewportHeight =
                viewport.clientHeight -
                (parseFloat(viewportStyle.paddingTop) || 0) -
                (parseFloat(viewportStyle.paddingBottom) || 0);
            const tableHeight = Math.max(viewportHeight, headerHeight + Math.max(contentHeight, bodyHeight));

            viewport.style.setProperty('--data-table-content-height', `${Math.max(contentHeight, viewportHeight)}px`);
            viewport.style.setProperty('--data-table-height', `${tableHeight}px`);
            viewport.style.setProperty(
                '--prevent-vanishing-sticky-header',
                `${Math.max(0, tableHeight - headerHeight - bodyHeight)}px`
            );
        };

        updateHeight();

        if (typeof ResizeObserver === 'undefined') return;

        const observer = new ResizeObserver(updateHeight);
        observer.observe(viewport);
        observer.observe(header);
        observer.observe(body);

        return () => observer.disconnect();
    }, [viewportRef, headerRef, bodyRef, contentHeight]);
};
