// Copyright 2023 Specter Ops, Inc.
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

import { fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Menu, MenuContent } from 'doodle-ui';
import * as hooks from '../../../hooks';
import { render } from '../../../test-utils';
import CopyMenuItem from './CopyMenuItem';

const useExploreSelectedItemSpy = vi.spyOn(hooks, 'useExploreSelectedItem');

// CopyMenuItem renders a MenuSub which requires a Radix Menu context.
// MenuContent forceMount ensures content is visible immediately in JSDOM
// (where CSS transitions never fire and Presence would otherwise block mount).
// modal={false} prevents the DismissableLayer from installing pointer capture on body,
// which would intercept clicks on the portal-rendered SubContent before onSelect fires.
const MenuWrapper = ({ children }: { children: React.ReactNode }) => (
    <Menu open modal={false}>
        <MenuContent forceMount>{children}</MenuContent>
    </Menu>
);

describe('CopyMenuItem', () => {
    const selectedNode = {
        label: 'foo',
    };

    const setup = () => {
        useExploreSelectedItemSpy.mockReturnValue({ selectedItemQuery: { data: selectedNode } } as any);
        const screen = render(
            <MenuWrapper>
                <CopyMenuItem />
            </MenuWrapper>
        );
        return screen;
    };

    it('handles copying the name', async () => {
        const screen = setup();

        const user = userEvent.setup();

        // Spy on clipboard.writeText rather than reading back from clipboard storage.
        // jsdom returns zero bounding rects for all elements, which breaks Radix's
        // "safe polygon" that keeps a submenu open while the pointer travels from the
        // SubTrigger into the SubContent. The submenu closes before a userEvent.click
        // can land on the item, so onSelect never fires. Using fireEvent.click instead
        // dispatches the click event directly — no pointer-travel required — so Radix
        // processes it and calls onSelect correctly.
        const clipboardSpy = vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue(undefined);

        const copyOption = screen.getByRole('menuitem', { name: /copy/i });
        await user.hover(copyOption);

        // Radix renders the submenu as role="menuitem" entries — no tooltip involved
        const nameOption = await screen.findByRole('menuitem', { name: /^name$/i });
        fireEvent.click(nameOption);

        expect(clipboardSpy).toHaveBeenCalledWith(selectedNode.label);
    });
});
