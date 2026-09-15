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

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { Button, IconButton } from './Button';

describe('Button', () => {
    it('defaults to type button', () => {
        render(<Button>Button</Button>);

        expect(screen.getByRole('button', { name: 'Button' }).getAttribute('type')).toBe('button');
    });

    it('accepts type submit', () => {
        render(<Button type='submit'>Submit</Button>);

        expect(screen.getByRole('button', { name: 'Submit' }).getAttribute('type')).toBe('submit');
    });
});

describe('IconButton', () => {
    it('renders an AppIcon and defaults the tooltip to its accessible label', async () => {
        const user = userEvent.setup();
        const { container } = render(
            <IconButton aria-label='More information'>
                <AppIcon.Info />
            </IconButton>
        );

        const button = screen.getByRole('button', { name: 'More information' });

        await user.hover(button);

        expect((await screen.findByRole('tooltip')).textContent).toBe('More information');
        expect(container.querySelector('svg')).not.toBeNull();
    });

    it('displays its tooltip when hovered', async () => {
        const user = userEvent.setup();
        render(
            <IconButton aria-label='More information' tooltip='Additional context'>
                <AppIcon.Info />
            </IconButton>
        );

        const button = screen.getByRole('button', { name: 'More information' });

        await user.hover(button);

        expect((await screen.findByRole('tooltip')).textContent).toBe('Additional context');
        expect(button.classList.contains('inline-grid')).toBe(true);
        expect(button.getAttribute('class')).not.toContain('(state) =>');
    });
});
