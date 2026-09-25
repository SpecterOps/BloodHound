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
import { Icon } from '../Icon';
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
        expect(container.querySelector('svg')?.getAttribute('aria-label')).toBeNull();
        expect(container.querySelector('svg')?.getAttribute('aria-hidden')).toBe('true');
    });

    it('supports tooltip text that differs from the accessible label', async () => {
        const user = userEvent.setup();
        render(
            <IconButton aria-label='Delete extension' tooltip='Built-in extensions cannot be deleted'>
                <AppIcon.Info />
            </IconButton>
        );

        await user.hover(screen.getByRole('button', { name: 'Delete extension' }));

        expect((await screen.findByRole('tooltip')).textContent).toBe('Built-in extensions cannot be deleted');
    });

    it('hides both the button and child Icon tooltips when requested', async () => {
        const user = userEvent.setup();
        render(
            <IconButton aria-label='More information' hideTooltip>
                <Icon aria-label='More information'>
                    <AppIcon.Info />
                </Icon>
            </IconButton>
        );

        await user.hover(screen.getByRole('button', { name: 'More information' }));

        expect(screen.queryByRole('tooltip')).toBeNull();
    });

    it('suppresses the child Icon tooltip in favor of the button tooltip', async () => {
        const user = userEvent.setup();
        render(
            <IconButton aria-label='Button information'>
                <Icon aria-label='Icon information'>
                    <AppIcon.Info />
                </Icon>
            </IconButton>
        );

        await user.hover(screen.getByRole('button', { name: 'Button information' }));

        expect((await screen.findByRole('tooltip')).textContent).toBe('Button information');
        expect(screen.queryByText('Icon information')).toBeNull();
    });

    it('uses the enabled button as the tooltip trigger on keyboard focus', async () => {
        const user = userEvent.setup();
        render(
            <IconButton aria-label='Keyboard information'>
                <AppIcon.Info />
            </IconButton>
        );

        const button = screen.getByRole('button', { name: 'Keyboard information' });

        await user.tab();

        expect(document.activeElement).toBe(button);
        expect((await screen.findByRole('tooltip')).textContent).toBe('Keyboard information');
    });

    it('uses a non-disabled wrapper as the tooltip trigger for disabled buttons', async () => {
        const user = userEvent.setup();
        render(
            <IconButton aria-label='Unavailable action' disabled>
                <AppIcon.Info />
            </IconButton>
        );

        const button = screen.getByRole('button', { name: 'Unavailable action' });
        const tooltipTrigger = button.parentElement;

        expect((button as HTMLButtonElement).disabled).toBe(true);
        expect(tooltipTrigger?.tagName).toBe('SPAN');
        expect(tooltipTrigger?.getAttribute('data-state')).toBe('closed');

        await user.hover(tooltipTrigger!);

        expect((await screen.findByRole('tooltip')).textContent).toBe('Unavailable action');
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
        expect(iconWrapper?.classList.contains('[&>*]:size-full')).toBe(true);
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

    it('defaults to the transparent default variant', () => {
        render(
            <IconButton aria-label='Default action'>
                <AppIcon.Info />
            </IconButton>
        );

        const button = screen.getByRole('button', { name: 'Default action' });

        expect(button.classList.contains('hover:text-primary')).toBe(true);
        expect(button.classList.contains('active:bg-transparent')).toBe(true);
        expect(button.classList.contains('bg-primary')).toBe(false);
        expect(button.classList.contains('bg-secondary-btn-fill')).toBe(false);
        expect(button.classList.contains('shadow-outer-1')).toBe(false);
    });

    it('uses className and the icon currentColor for color', () => {
        const { container } = render(
            <IconButton aria-label='Delete' className='text-status-error-main'>
                <AppIcon.FilterOutline />
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
