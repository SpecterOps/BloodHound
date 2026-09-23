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

const renderDrawer = () =>
    render(
        <Drawer>
            <DrawerTrigger>Open drawer</DrawerTrigger>
            <DrawerContent>
                <DrawerHeader>
                    <DrawerTitle>Collection plan</DrawerTitle>
                    <DrawerClose>Close drawer</DrawerClose>
                </DrawerHeader>
                <DrawerBody>Drawer content</DrawerBody>
            </DrawerContent>
        </Drawer>
    );

describe('Drawer', () => {
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
