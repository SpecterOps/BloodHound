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
import { primaryClasses, secondaryClasses } from '../Button/Button.styles';
import { IconButton } from './IconButton';

describe('IconButton', () => {
    it('renders an AppIcon and defaults the tooltip to its accessible label', async () => {
        const user = userEvent.setup();
        const { container } = render(
            <IconButton aria-label='More information'>
                <AppIcon.Info />
            </IconButton>
        );

        await user.hover(container.querySelector('svg') as SVGSVGElement);

        expect((await screen.findByRole('tooltip')).textContent).toBe('More information');
        expect(container.querySelector('svg')?.getAttribute('aria-label')).toBe('More information');
    });

    it('applies its styles to the button', () => {
        const { container } = render(
            <IconButton aria-label='More information'>
                <AppIcon.Info />
            </IconButton>
        );

        const button = screen.getByRole('button', { name: 'More information' });
        const iconWrapper = container.querySelector('svg')?.parentElement;

        expect(button.classList.contains('inline-grid')).toBe(true);
        expect(button.getAttribute('class')).not.toContain('(state) =>');
        expect(iconWrapper?.classList.contains('size-[var(--icon-button-icon-size)]')).toBe(true);
        expect(iconWrapper?.classList.contains('items-center')).toBe(true);
    });

    it.each([
        ['primary', primaryClasses],
        ['secondary', secondaryClasses],
    ] as const)('uses the same %s variant styles as Button', (variant, variantClasses) => {
        render(
            <IconButton aria-label={`${variant} action`} variant={variant}>
                <AppIcon.Info />
            </IconButton>
        );

        const button = screen.getByRole('button', { name: `${variant} action` });
        const expectedClasses = variantClasses.flatMap((classNames) => classNames.split(' '));

        expect(expectedClasses.every((className) => button.classList.contains(className))).toBe(true);
    });

    it('defaults to the primary Button variant', () => {
        render(
            <IconButton aria-label='Primary action'>
                <AppIcon.Info />
            </IconButton>
        );

        expect(screen.getByRole('button', { name: 'Primary action' }).classList.contains('bg-primary')).toBe(true);
    });

    it('uses className and the icon currentColor for color', () => {
        const { container } = render(
            <IconButton aria-label='Delete' className='text-status-error-main'>
                <AppIcon.FileMagnifyingGlass />
            </IconButton>
        );

        const button = screen.getByRole('button', { name: 'Delete' });
        const iconPaths = container.querySelectorAll('svg path');

        expect(button.classList.contains('text-status-error-main')).toBe(true);
        expect((button as HTMLButtonElement).style.color).toBe('');
        expect(Array.from(iconPaths).every((path) => path.getAttribute('fill') === 'currentColor')).toBe(true);
        expect(Array.from(iconPaths).every((path) => path.getAttribute('stroke') === 'currentColor')).toBe(true);
    });
});
