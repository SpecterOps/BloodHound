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
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { LoadingSpinner } from './LoadingSpinner';

expect.extend(matchers);

describe('LoadingSpinner', () => {
    it('exposes a named indeterminate progress indicator without announcing the decorative icon', () => {
        render(<LoadingSpinner />);

        expect(screen.getByRole('progressbar', { name: 'Loading' })).not.toHaveAttribute('aria-valuenow');
        expect(screen.queryByRole('img')).not.toBeInTheDocument();
    });

    it('applies inline styling, accessibility attributes, and event handlers to the same root', async () => {
        const user = userEvent.setup();
        const onClick = vi.fn();

        render(
            <LoadingSpinner
                id='save-progress'
                aria-label='Saving changes'
                className='size-6 text-secondary'
                style={{ marginLeft: 8 }}
                onClick={onClick}
            />
        );

        const spinner = screen.getByRole('progressbar', { name: 'Saving changes' });
        const icon = spinner.querySelector('svg');

        expect(spinner).toHaveAttribute('id', 'save-progress');
        expect(spinner).toHaveClass('size-6', 'text-secondary');
        expect(spinner).toHaveStyle({ marginLeft: '8px' });
        expect(icon).toBeInTheDocument();
        expect(icon).not.toHaveClass('size-6', 'text-secondary');

        await user.click(spinner);

        expect(onClick).toHaveBeenCalledOnce();
    });

    it('styles the overlay backdrop and icon independently without leaking component props to the DOM', () => {
        render(<LoadingSpinner variant='overlay' className='bg-black/10' spinnerClassName='size-6 text-secondary' />);

        const overlay = screen.getByRole('progressbar', { name: 'Loading' });
        const icon = overlay.querySelector('svg');

        expect(overlay).toHaveClass('bg-black/10');
        expect(overlay).not.toHaveClass('size-6', 'text-secondary');
        expect(icon).toHaveClass('size-6', 'text-secondary');
        expect(icon).not.toHaveClass('bg-black/10');
        expect(overlay).not.toHaveAttribute('variant');
        expect(overlay).not.toHaveAttribute('spinnerClassName');
    });
});
