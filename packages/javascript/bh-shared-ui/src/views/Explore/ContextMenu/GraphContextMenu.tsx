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

import { Menu, MenuContent, MenuTrigger } from 'doodle-ui';
import { FC, ReactNode, useCallback, useEffect, useRef } from 'react';

// Shared canvas integration for the legacy and privilege-zone graph menus.
const GraphContextMenu: FC<{
    contextMenu: { mouseX: number; mouseY: number } | null;
    onClose: () => void;
    children: ReactNode;
}> = ({ contextMenu, onClose, children }) => {
    const lastPosition = useRef(contextMenu);
    useEffect(() => {
        if (contextMenu) lastPosition.current = contextMenu;
    }, [contextMenu]);
    const position = contextMenu ?? lastPosition.current;
    const returnFocusRef = useRef<HTMLElement | null>(null);
    const captureReturnFocus = useCallback((content: HTMLDivElement | null) => {
        if (content && !content.contains(document.activeElement))
            returnFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    }, []);
    return (
        <Menu
            // Confirmation dialogs own focus while open.
            modal={false}
            open={contextMenu !== null}
            onOpenChange={(open) => {
                if (!open) onClose();
            }}>
            <MenuTrigger asChild>
                <span
                    aria-hidden
                    tabIndex={-1}
                    className='fixed pointer-events-none'
                    style={{ left: position?.mouseX ?? 0, top: position?.mouseY ?? 0 }}
                />
            </MenuTrigger>
            <MenuContent
                className='max-h-[var(--radix-dropdown-menu-content-available-height)] overflow-y-auto'
                align='start'
                sideOffset={0}
                // Graph engines restore canvas focus after opening the menu. Outside clicks and Escape still dismiss it.
                onFocusOutside={(event) => event.preventDefault()}
                onPointerUpCapture={(event) => {
                    // The opening context click can release over a newly mounted menu item.
                    // Prevent Radix from turning that release into a synthetic item click.
                    if (event.button !== 0 || event.ctrlKey) event.preventDefault();
                }}
                ref={captureReturnFocus}
                onCloseAutoFocus={(event) => {
                    event.preventDefault();
                    if (!document.activeElement?.closest('[role="dialog"]')) returnFocusRef.current?.focus();
                }}
                onInteractOutside={(event) => {
                    if ((event.target as HTMLElement).closest('[role="dialog"]')) event.preventDefault();
                }}>
                {children}
            </MenuContent>
        </Menu>
    );
};

export default GraphContextMenu;
