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
import { createRef } from 'react';
import { Tooltip } from './Tooltip';

describe('Tooltip', () => {
    it('uses AppIcon.Info as the default trigger', () => {
        const { container } = render(<Tooltip tooltip='Helpful context' />);

        expect(screen.queryByRole('button')).toBeNull();
        const trigger = screen.getByRole('img');

        expect(trigger.getAttribute('tabindex')).toBe('0');
        expect(trigger.classList.contains('rounded-full')).toBe(true);
        expect(container.querySelector('svg')).not.toBeNull();
        expect(container.querySelector('svg')?.getAttribute('aria-hidden')).toBe('true');
        expect(container.querySelector('svg')?.getAttribute('viewBox')).toBe('0 0 24 24');
        expect(container.querySelector('svg')?.getAttribute('width')).toBe('16');
        expect(container.querySelector('svg')?.getAttribute('height')).toBe('16');
    });

    it('does not submit a parent form when the default trigger is clicked', async () => {
        const user = userEvent.setup();
        const handleSubmit = vi.fn((event: React.FormEvent) => event.preventDefault());

        render(
            <form onSubmit={handleSubmit}>
                <Tooltip tooltip='Helpful context' />
            </form>
        );

        await user.click(screen.getByRole('img'));

        expect(handleSubmit).not.toHaveBeenCalled();
    });

    it('preserves an explicit accessible name for the default trigger', () => {
        render(<Tooltip tooltip='Helpful context' triggerProps={{ 'aria-label': 'More details' }} />);

        expect(screen.getByRole('img').getAttribute('aria-label')).toBe('More details');
    });

    it('opens the default trigger tooltip on keyboard focus', async () => {
        const user = userEvent.setup();
        render(<Tooltip tooltip='Helpful context' />);

        await user.tab();

        const trigger = screen.getByRole('img');
        expect(document.activeElement).toBe(trigger);

        const tooltip = await screen.findByRole('tooltip', { name: 'Helpful context' });
        expect(trigger.getAttribute('aria-describedby')).toBe(tooltip.id);
        expect(trigger.getAttribute('aria-label')).toBe('Information');
    });

    it('applies the default overlay z-index', async () => {
        const user = userEvent.setup();
        render(
            <Tooltip tooltip='Helpful context'>
                <button>Show tooltip</button>
            </Tooltip>
        );

        await user.hover(screen.getByRole('button', { name: 'Show tooltip' }));

        const tooltipContent = (await screen.findByRole('tooltip')).parentElement;

        expect(tooltipContent?.classList.contains('z-[1700]')).toBe(true);
        expect(tooltipContent?.classList.contains('border-[#939597]')).toBe(true);
        expect(tooltipContent?.classList.contains('max-w-[225px]')).toBe(true);
    });

    it('applies a content id without replacing the trigger description id', async () => {
        render(<Tooltip open tooltip='Helpful context' contentProps={{ id: 'tooltip-popup' }} />);

        const trigger = screen.getByRole('img');
        const tooltip = await screen.findByRole('tooltip', { name: 'Helpful context' });

        expect(tooltip.parentElement?.getAttribute('id')).toBe('tooltip-popup');
        expect(tooltip.id).not.toBe('tooltip-popup');
        expect(trigger.getAttribute('aria-describedby')).toBe(tooltip.id);
    });

    it.each(['top', 'bottom', 'left', 'right'] as const)(
        'positions the tooltip on the %s of its trigger',
        async (tooltipPosition) => {
            render(
                <Tooltip
                    open
                    tooltip='Helpful context'
                    tooltipPosition={tooltipPosition}
                    contentProps={{ avoidCollisions: false }}>
                    <button>Show tooltip</button>
                </Tooltip>
            );

            const tooltipContent = (await screen.findByRole('tooltip')).parentElement;

            expect(tooltipContent?.getAttribute('data-side')).toBe(tooltipPosition);
        }
    );

    it('forwards refs to the trigger', () => {
        const ref = createRef<HTMLButtonElement>();

        render(
            <Tooltip ref={ref} tooltip='Helpful context'>
                <button>Show tooltip</button>
            </Tooltip>
        );

        expect(ref.current).toBe(screen.getByRole('button', { name: 'Show tooltip' }));
    });
});
