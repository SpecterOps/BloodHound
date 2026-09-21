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
import { faDiagramProject } from '@fortawesome/free-solid-svg-icons';
import userEvent from '@testing-library/user-event';
import { MenuItem } from 'doodle-ui';
import { act, render, screen } from '../../test-utils';
import GraphMenu from './GraphMenu';

describe('GraphMenu', () => {
    const onSelectFn = vi.fn();

    afterEach(() => {
        onSelectFn.mockClear();
    });

    const setup = ({ tooltip }: { tooltip?: string } = {}) => {
        render(
            <GraphMenu label='Layout' icon={faDiagramProject} tooltip={tooltip}>
                <MenuItem onSelect={onSelectFn}>First option</MenuItem>
                <MenuItem>Second option</MenuItem>
            </GraphMenu>
        );

        const user = userEvent.setup();

        return { user };
    };

    it('renders an accessible trigger button labeled by the label prop', () => {
        setup();

        const trigger = screen.getByRole('button', { name: 'Layout' });

        expect(trigger).toBeInTheDocument();
        expect(trigger).toHaveAttribute('aria-haspopup', 'menu');
        expect(trigger).toHaveAttribute('aria-expanded', 'false');
    });

    it('shows a tooltip on hover, defaulting to the label when no tooltip is provided', async () => {
        const { user } = setup();

        await user.hover(screen.getByRole('button', { name: 'Layout' }));

        expect(await screen.findByRole('tooltip', { name: 'Layout' })).toBeVisible();
    });

    it('shows the provided tooltip text when the tooltip prop is set', async () => {
        const { user } = setup({ tooltip: 'Change layout' });

        await user.hover(screen.getByRole('button', { name: 'Layout' }));

        expect(await screen.findByRole('tooltip', { name: 'Change layout' })).toBeVisible();
    });

    it('opens the menu and renders its children when the trigger is clicked', async () => {
        const { user } = setup();
        const trigger = screen.getByRole('button', { name: 'Layout' });

        await user.click(trigger);

        const menu = await screen.findByRole('menu');
        expect(menu).toBeVisible();
        expect(trigger).toHaveAttribute('aria-expanded', 'true');
        expect(await screen.findByRole('menuitem', { name: 'First option' })).toBeInTheDocument();
        expect(await screen.findByRole('menuitem', { name: 'Second option' })).toBeInTheDocument();
    });

    it('invokes onSelect, closes the menu, and restores focus to the trigger when an item is chosen', async () => {
        const { user } = setup();
        const trigger = screen.getByRole('button', { name: 'Layout' });

        await user.click(trigger);
        await user.click(await screen.findByRole('menuitem', { name: 'First option' }));

        expect(onSelectFn).toHaveBeenCalledOnce();
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
        expect(trigger).toHaveAttribute('aria-expanded', 'false');
        expect(trigger).toHaveFocus();
    });

    it('supports keyboard open and close, restoring focus to the trigger', async () => {
        const { user } = setup();
        const trigger = screen.getByRole('button', { name: 'Layout' });

        act(() => trigger.focus());
        await user.keyboard('{Enter}');

        expect(await screen.findByRole('menu')).toBeVisible();
        expect(trigger).toHaveAttribute('aria-expanded', 'true');

        await user.keyboard('{Escape}');

        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
        expect(trigger).toHaveAttribute('aria-expanded', 'false');
        expect(trigger).toHaveFocus();
    });
});
