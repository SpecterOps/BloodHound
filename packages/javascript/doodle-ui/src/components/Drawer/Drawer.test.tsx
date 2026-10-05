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
import { userEvent, type UserEvent } from '@testing-library/user-event';
import { expect } from 'vitest';
import { Drawer, DrawerBody, DrawerClose, DrawerContent, DrawerHeader, DrawerTitle, DrawerTrigger } from './Drawer';

expect.extend(matchers);

const renderDrawer = ({
    className,
    ...drawerProps
}: React.ComponentProps<typeof Drawer> & { className?: string } = {}) =>
    render(
        <Drawer {...drawerProps}>
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
    let user: UserEvent;

    beforeEach(() => {
        user = userEvent.setup();
    });

    it('defaults to a compact right-side drawer', async () => {
        renderDrawer();

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        const dialog = screen.getByRole('dialog', { name: 'Collection plan' });
        expect(dialog).toHaveAttribute('data-swipe-direction', 'right');
        expect(dialog).toHaveClass('data-[swipe-axis=x]:sm:[--drawer-content-width:24rem]');
        expect(dialog).toHaveClass('data-[swipe-axis=x]:[--drawer-content-width:75%]');
    });

    it('allows the side drawer width to be overridden', async () => {
        renderDrawer({ className: 'w-[860px]' });

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        const dialog = screen.getByRole('dialog', { name: 'Collection plan' });
        expect(dialog).toHaveAttribute('data-swipe-direction', 'right');
        expect(dialog).toHaveClass('w-[860px]');
        expect(dialog).not.toHaveClass('w-[var(--drawer-content-width,auto)]');
    });

    it.each([
        ['left', 'data-[swipe-direction=left]:left-0'],
        ['up', 'data-[swipe-direction=up]:top-0'],
        ['down', 'data-[swipe-direction=down]:bottom-0'],
    ] as const)('opens from the %s', async (swipeDirection, positionClass) => {
        renderDrawer({ swipeDirection });

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        const dialog = screen.getByRole('dialog', { name: 'Collection plan' });
        expect(dialog).toHaveAttribute('data-swipe-direction', swipeDirection);
        expect(dialog).toHaveClass(positionClass);
    });

    it('opens from its trigger and closes when the backdrop is clicked', async () => {
        renderDrawer();

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        const dialog = screen.getByRole('dialog', { name: 'Collection plan' });
        expect(dialog).toBeInTheDocument();
        expect(screen.getByText('Drawer content')).toBeInTheDocument();

        const overlay = document.querySelector('[data-slot="drawer-overlay"]');
        expect(overlay).toBeInstanceOf(HTMLElement);
        await user.click(overlay as HTMLElement);

        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('renders a swipe handle when requested', async () => {
        renderDrawer({ showSwipeHandle: true });

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        expect(document.querySelector('[data-slot="drawer-swipe-handle"]')).toBeInTheDocument();
    });

    it('omits the overlay for a non-modal drawer', async () => {
        renderDrawer({ modal: false });

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        expect(screen.getByRole('dialog', { name: 'Collection plan' })).toBeInTheDocument();
        expect(document.querySelector('[data-slot="drawer-overlay"]')).not.toBeInTheDocument();
    });

    it('closes with its close button', async () => {
        renderDrawer();

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        await user.click(screen.getByRole('button', { name: 'Close drawer' }));
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('closes with the Escape key', async () => {
        renderDrawer();

        await user.click(screen.getByRole('button', { name: 'Open drawer' }));
        await user.keyboard('{Escape}');
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });
});
