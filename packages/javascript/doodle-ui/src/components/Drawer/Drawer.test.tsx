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
import '@testing-library/jest-dom';
import matchers from '@testing-library/jest-dom/matchers';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect } from 'vitest';
import { Drawer, DrawerBody, DrawerClose, DrawerContent, DrawerHeader, DrawerTitle, DrawerTrigger } from './Drawer';

expect.extend(matchers);

const renderDrawer = ({
    swipeDirection,
    className,
}: {
    swipeDirection?: React.ComponentProps<typeof Drawer>['swipeDirection'];
    className?: string;
} = {}) =>
    render(
        <Drawer swipeDirection={swipeDirection}>
            <DrawerTrigger>Open drawer</DrawerTrigger>
            <DrawerContent className={className}>
                <DrawerHeader>
                    <DrawerTitle>Collection plan</DrawerTitle>
                    <DrawerClose>Close drawer</DrawerClose>
                </DrawerHeader>
                <DrawerBody>Drawer content</DrawerBody>
            </DrawerContent>
        </Drawer>
    );

describe('Drawer', () => {
    it('defaults to a compact right-side drawer', async () => {
        const user = userEvent.setup();
        renderDrawer();

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        const dialog = screen.getByRole('dialog', { name: 'Collection plan' });
        expect(dialog).toHaveAttribute('data-swipe-direction', 'right');
        expect(dialog).toHaveClass('max-w-sm');
    });

    it('allows the side drawer width to be overridden', async () => {
        const user = userEvent.setup();
        renderDrawer({ className: 'max-w-[860px]' });

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        const dialog = screen.getByRole('dialog', { name: 'Collection plan' });
        expect(dialog).toHaveAttribute('data-swipe-direction', 'right');
        expect(dialog).toHaveClass('max-w-[860px]');
        expect(dialog).not.toHaveClass('max-w-sm');
    });

    it.each([
        ['left', 'mr-auto'],
        ['up', 'self-start'],
        ['down', 'self-end'],
    ] as const)('opens from the %s', async (swipeDirection, positionClass) => {
        const user = userEvent.setup();
        renderDrawer({ swipeDirection });

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        const dialog = screen.getByRole('dialog', { name: 'Collection plan' });
        expect(dialog).toHaveAttribute('data-swipe-direction', swipeDirection);
        expect(dialog).toHaveClass(positionClass);
    });

    it('opens from its trigger and closes when the backdrop is clicked', async () => {
        const user = userEvent.setup();
        renderDrawer();

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        const dialog = screen.getByRole('dialog', { name: 'Collection plan' });
        expect(dialog).toBeInTheDocument();
        expect(screen.getByText('Drawer content')).toBeInTheDocument();

        const backdrop = dialog.parentElement?.previousElementSibling;
        expect(backdrop).toBeInstanceOf(HTMLElement);
        await user.click(backdrop as HTMLElement);

        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('closes with its close button and the Escape key', async () => {
        const user = userEvent.setup();
        renderDrawer();

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        await user.click(screen.getByRole('button', { name: 'Close drawer' }));
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        await user.keyboard('{Escape}');
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });
});
