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
import { expect } from 'vitest';
import { RadioGroup, RadioItem } from './RadioGroup';

expect.extend(matchers);

// The Radix UI radio <button> element is not programmatically linked to its adjacent <label>
// via for/id, so accessible-name queries like getByRole('radio', { name: '...' }) do not work.
// Instead, radio buttons are queried directly by their value attribute.
const getRadioByValue = (container: HTMLElement, value: string): HTMLElement =>
    container.querySelector<HTMLElement>(`[role="radio"][value="${value}"]`)!;

describe('RadioGroup', () => {
    it('renders children', () => {
        render(
            <RadioGroup>
                <RadioItem value='a' label='Option A' />
                <RadioItem value='b' label='Option B' />
            </RadioGroup>
        );

        expect(screen.getByText('Option A')).toBeInTheDocument();
        expect(screen.getByText('Option B')).toBeInTheDocument();
    });

    it('applies flex class when row prop is true', () => {
        const { container } = render(
            <RadioGroup row>
                <RadioItem value='a' label='Option A' />
            </RadioGroup>
        );

        expect(container.firstChild).toHaveClass('flex');
    });

    it('does not apply flex class when row prop is false', () => {
        const { container } = render(
            <RadioGroup row={false}>
                <RadioItem value='a' label='Option A' />
            </RadioGroup>
        );

        expect(container.firstChild).not.toHaveClass('flex');
    });

    it('does not apply flex class when row prop is omitted', () => {
        const { container } = render(
            <RadioGroup>
                <RadioItem value='a' label='Option A' />
            </RadioGroup>
        );

        expect(container.firstChild).not.toHaveClass('flex');
    });

    it('merges additional className prop', () => {
        const { container } = render(
            <RadioGroup className='custom-class'>
                <RadioItem value='a' label='Option A' />
            </RadioGroup>
        );

        expect(container.firstChild).toHaveClass('custom-class');
    });

    it('calls onValueChange when a radio item is clicked', async () => {
        const user = userEvent.setup();
        const onValueChange = vi.fn();

        const { container } = render(
            <RadioGroup onValueChange={onValueChange}>
                <RadioItem value='a' label='Option A' />
                <RadioItem value='b' label='Option B' />
            </RadioGroup>
        );

        await user.click(getRadioByValue(container, 'a'));
        expect(onValueChange).toHaveBeenCalledWith('a');

        await user.click(getRadioByValue(container, 'b'));
        expect(onValueChange).toHaveBeenCalledWith('b');
    });

    it('reflects the controlled value via aria-checked', () => {
        const { container } = render(
            <RadioGroup value='b'>
                <RadioItem value='a' label='Option A' />
                <RadioItem value='b' label='Option B' />
            </RadioGroup>
        );

        expect(getRadioByValue(container, 'b')).toHaveAttribute('aria-checked', 'true');
        expect(getRadioByValue(container, 'a')).toHaveAttribute('aria-checked', 'false');
    });
});

describe('RadioItem', () => {
    it('renders the label text', () => {
        render(
            <RadioGroup>
                <RadioItem value='a' label='My Label' />
            </RadioGroup>
        );

        expect(screen.getByText('My Label')).toBeInTheDocument();
    });

    it('renders a radio button with the correct value attribute', () => {
        const { container } = render(
            <RadioGroup>
                <RadioItem value='test-value' label='Test' />
            </RadioGroup>
        );

        expect(getRadioByValue(container, 'test-value')).toHaveAttribute('value', 'test-value');
    });

    it('renders as disabled when disabled prop is provided', () => {
        const { container } = render(
            <RadioGroup>
                <RadioItem value='a' label='Disabled Option' disabled />
            </RadioGroup>
        );

        expect(getRadioByValue(container, 'a')).toBeDisabled();
    });

    it('does not call onValueChange when a disabled item is clicked', async () => {
        const user = userEvent.setup();
        const onValueChange = vi.fn();

        const { container } = render(
            <RadioGroup onValueChange={onValueChange}>
                <RadioItem value='a' label='Disabled Option' disabled />
            </RadioGroup>
        );

        await user.click(getRadioByValue(container, 'a'));
        expect(onValueChange).not.toHaveBeenCalled();
    });
});
