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
import { expect, vi } from 'vitest';
import { Checkbox, CheckboxWithLabel } from './Checkbox';

expect.extend(matchers);

describe('Checkbox Tests', () => {
    it('renders an uncontrolled indeterminate checkbox', () => {
        render(<Checkbox aria-label='Parent checkbox' defaultChecked='indeterminate' />);

        const checkbox = screen.getByRole('checkbox', { name: 'Parent checkbox' });

        expect(checkbox).toHaveAttribute('data-state', 'indeterminate');
        expect(checkbox).toHaveAttribute('aria-checked', 'mixed');
    });

    it('does not call onCheckedChange when disabled', async () => {
        const user = userEvent.setup();
        const onCheckedChange = vi.fn();

        render(<Checkbox aria-label='Disabled checkbox' disabled onCheckedChange={onCheckedChange} />);

        await user.click(screen.getByRole('checkbox', { name: 'Disabled checkbox' }));

        expect(onCheckedChange).not.toHaveBeenCalled();
    });
});

describe('CheckboxWithLabel Tests', () => {
    it('associates the label with the checkbox', () => {
        render(<CheckboxWithLabel label='All zones' />);

        expect(screen.getByRole('checkbox', { name: 'All zones' })).toBeInTheDocument();
    });

    it('toggles when the label is clicked', async () => {
        const user = userEvent.setup();

        render(<CheckboxWithLabel label='All zones' defaultChecked={false} />);

        const checkbox = screen.getByRole('checkbox', { name: 'All zones' });

        await user.click(screen.getByText('All zones'));

        expect(checkbox).toHaveAttribute('aria-checked', 'true');
    });

    it('applies error state to checkbox and label', () => {
        render(<CheckboxWithLabel label='Required option' error />);

        expect(screen.getByRole('checkbox', { name: 'Required option' })).toHaveAttribute('aria-invalid', 'true');
        expect(screen.getByText('Required option')).toHaveClass('text-error');
    });

    it('does not toggle when disabled label is clicked', async () => {
        const user = userEvent.setup();
        const onCheckedChange = vi.fn();

        render(<CheckboxWithLabel label='Disabled option' disabled onCheckedChange={onCheckedChange} />);

        await user.click(screen.getByText('Disabled option'));

        expect(onCheckedChange).not.toHaveBeenCalled();
    });
});
