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

import { Field } from '@base-ui/react/field';
import { act, fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, vi } from 'vitest';
import { Slider } from './Slider';

describe('Slider Tests', () => {
    it('renders a thumb exposing the provided aria-label', () => {
        render(<Slider defaultValue={40} thumbAriaLabel='Horizontal spacing' />);

        expect(screen.getByRole('slider', { name: 'Horizontal spacing' })).toBeInTheDocument();
    });

    it('applies aria-describedby to the slider input', () => {
        render(
            <>
                <p id='slider-description'>Choose a value</p>
                <Slider defaultValue={50} thumbAriaLabel='Value' aria-describedby='slider-description' />
            </>
        );

        expect(screen.getByRole('slider', { name: 'Value' })).toHaveAttribute('aria-describedby', 'slider-description');
    });

    it('calls onValueChange when the value is changed via the keyboard', async () => {
        const user = userEvent.setup();
        const onValueChange = vi.fn();

        render(
            <Slider defaultValue={50} min={0} max={100} step={1} thumbAriaLabel='Value' onValueChange={onValueChange} />
        );

        act(() => {
            const thumb = screen.getByRole('slider', { name: 'Value' });
            thumb.focus();
        });
        await user.keyboard('{ArrowRight}');

        expect(onValueChange).toHaveBeenCalled();
    });

    it('does not call onValueChange when disabled', async () => {
        const user = userEvent.setup();
        const onValueChange = vi.fn();

        render(<Slider defaultValue={50} disabled thumbAriaLabel='Value' onValueChange={onValueChange} />);

        const thumb = screen.getByRole('slider', { name: 'Value' });
        thumb.focus();
        await user.keyboard('{ArrowRight}');

        expect(onValueChange).not.toHaveBeenCalled();
    });

    it('keeps controlled styling aligned when the parent does not accept a value change', async () => {
        const user = userEvent.setup();

        render(<Slider value={0} min={0} max={100} thumbAriaLabel='Value' onValueChange={vi.fn()} />);

        const thumbInput = screen.getByRole('slider', { name: 'Value' });
        const thumbDot = thumbInput.parentElement?.querySelector('span[aria-hidden]');

        await act(async () => {
            thumbInput.focus();
            await user.keyboard('{ArrowRight}');
        });

        expect(thumbInput).toHaveValue('0');
        expect(thumbDot).toHaveClass('hidden');
        expect(thumbDot).not.toHaveClass('block');
    });

    it('does not update uncontrolled styling when a value change is canceled', async () => {
        const user = userEvent.setup();

        render(
            <Slider
                defaultValue={0}
                min={0}
                max={100}
                step={1}
                thumbAriaLabel='Value'
                onValueChange={(_nextValue, eventDetails) => eventDetails.cancel()}
            />
        );

        const thumbInput = screen.getByRole('slider', { name: 'Value' });
        const thumbDot = thumbInput.parentElement?.querySelector('span[aria-hidden]');

        await act(async () => {
            thumbInput.focus();
            await user.keyboard('{ArrowRight}');
        });

        expect(thumbInput).toHaveValue('0');
        expect(thumbDot).toHaveClass('hidden');
        expect(thumbDot).not.toHaveClass('block');
    });

    it('shows the thumb dot when the value is greater than the minimum', () => {
        render(<Slider defaultValue={50} min={0} max={100} thumbAriaLabel='Value' />);

        const thumbInput = screen.getByRole('slider', { name: 'Value' });
        const thumbDot = thumbInput.parentElement?.querySelector('span[aria-hidden]');

        expect(thumbDot).toHaveClass('block');
        expect(thumbDot).not.toHaveClass('hidden');
    });

    it('applies a state-driven className callback to the root when disabled', () => {
        const { container } = render(
            <Slider defaultValue={50} disabled className={(state) => (state.disabled ? 'my-disabled-slider' : '')} />
        );

        expect(container.firstChild).toHaveClass('my-disabled-slider');
    });

    it('does not apply the disabled className from the callback when enabled', () => {
        const { container } = render(
            <Slider defaultValue={50} className={(state) => (state.disabled ? 'my-disabled-slider' : '')} />
        );

        expect(container.firstChild).not.toHaveClass('my-disabled-slider');
    });
});

describe('Slider ticks and value field', () => {
    it.each([undefined, false])(
        'inherits Field.Root disabled even when the Slider disabled prop is %s',
        async (disabled) => {
            const user = userEvent.setup();
            const onValueChange = vi.fn();
            const onValueCommitted = vi.fn();
            render(
                <Field.Root disabled>
                    <Slider
                        defaultValue={25}
                        disabled={disabled}
                        showValue
                        thumbAriaLabel='Value'
                        onValueChange={onValueChange}
                        onValueCommitted={onValueCommitted}
                    />
                </Field.Root>
            );
            const input = screen.getByRole('textbox');

            expect(input).toBeDisabled();
            expect(screen.getByRole('slider')).toBeDisabled();

            await user.type(input, '75{Enter}');
            fireEvent.blur(input);

            expect(input).toHaveValue('25');
            expect(screen.getByRole('slider')).toHaveValue('25');
            expect(onValueChange).not.toHaveBeenCalled();
            expect(onValueCommitted).not.toHaveBeenCalled();
        }
    );

    it('blocks pending edits when the parent field becomes disabled and supports re-enabling it', async () => {
        const user = userEvent.setup();
        const onValueChange = vi.fn();
        const onValueCommitted = vi.fn();
        const renderSlider = (disabled: boolean) => (
            <Field.Root disabled={disabled}>
                <Slider
                    defaultValue={25}
                    showValue
                    thumbAriaLabel='Value'
                    onValueChange={onValueChange}
                    onValueCommitted={onValueCommitted}
                />
            </Field.Root>
        );
        const { rerender } = render(renderSlider(false));
        const input = screen.getByRole('textbox');

        await user.clear(input);
        await user.type(input, '75');
        rerender(renderSlider(true));
        fireEvent.keyDown(input, { key: 'Enter' });
        fireEvent.blur(input);

        expect(input).toBeDisabled();
        expect(screen.getByRole('slider')).toHaveValue('25');
        expect(onValueChange).not.toHaveBeenCalled();
        expect(onValueCommitted).not.toHaveBeenCalled();

        rerender(renderSlider(false));
        expect(input).toBeEnabled();
        await user.click(input);
        await user.keyboard('{Enter}');

        expect(screen.getByRole('slider')).toHaveValue('75');
        expect(onValueChange).toHaveBeenCalledOnce();
        expect(onValueCommitted).toHaveBeenCalledOnce();
    });

    it('describes the default value field with its percent suffix', () => {
        render(<Slider showValue thumbAriaLabel='Opacity' />);

        expect(screen.getByRole('textbox', { name: 'Opacity value' })).toHaveAccessibleDescription('%');
        expect(screen.getByRole('slider', { name: 'Opacity' })).not.toHaveAttribute('aria-describedby');
    });

    it('preserves the custom field label and combines caller descriptions, errors, and suffix', async () => {
        const user = userEvent.setup();
        render(
            <>
                <p id='opacity-help'>Choose opacity.</p>
                <p id='opacity-detail'>Applies to the graph.</p>
                <Slider showValue valueAriaLabel='Graph opacity' aria-describedby='opacity-help opacity-detail' />
            </>
        );
        const input = screen.getByRole('textbox', { name: 'Graph opacity' });

        expect(input).toHaveAccessibleDescription('Choose opacity. Applies to the graph. %');

        await user.clear(input);
        await user.type(input, '101{Enter}');

        expect(input).toHaveAccessibleName('Graph opacity');
        expect(input).toHaveAccessibleDescription(
            'Choose opacity. Applies to the graph. Enter a value between 0 and 100. %'
        );
    });

    it('updates the description for custom suffixes and removes it when the suffix is empty', () => {
        const { rerender } = render(<Slider showValue valueSuffix='px' />);
        const input = screen.getByRole('textbox');

        expect(input).toHaveAccessibleDescription('px');

        rerender(<Slider showValue valueSuffix='' />);

        expect(input).not.toHaveAttribute('aria-describedby');
    });

    it.each([{ isComposing: true }, { keyCode: 229 }])(
        'ignores IME confirmation Enter with %j without preventing it or clearing the draft',
        async (compositionState) => {
            const user = userEvent.setup();
            const onValueChange = vi.fn();
            const onValueCommitted = vi.fn();
            render(
                <Slider
                    defaultValue={25}
                    showValue
                    thumbAriaLabel='Value'
                    onValueChange={onValueChange}
                    onValueCommitted={onValueCommitted}
                />
            );
            const input = screen.getByRole('textbox');

            await user.clear(input);
            await user.type(input, '75');

            expect(fireEvent.keyDown(input, { key: 'Enter', ...compositionState })).toBe(true);
            expect(input).toHaveValue('75');
            expect(screen.getByRole('slider')).toHaveValue('25');
            expect(onValueChange).not.toHaveBeenCalled();
            expect(onValueCommitted).not.toHaveBeenCalled();

            expect(fireEvent.keyDown(input, { key: 'Enter' })).toBe(false);
            expect(screen.getByRole('slider')).toHaveValue('75');
            expect(onValueCommitted).toHaveBeenCalledOnce();
        }
    );

    it.each([0, -1, NaN, Infinity, -Infinity])(
        'normalizes invalid step %s for field snapping and keyboard movement',
        async (step) => {
            const user = userEvent.setup();
            const onValueChange = vi.fn();
            const onValueCommitted = vi.fn();
            render(
                <Slider
                    step={step}
                    defaultValue={25}
                    showValue
                    thumbAriaLabel='Value'
                    onValueChange={onValueChange}
                    onValueCommitted={onValueCommitted}
                />
            );
            const input = screen.getByRole('textbox');
            const thumb = screen.getByRole('slider');

            await user.clear(input);
            await user.type(input, '26.4{Enter}');

            expect(input).toHaveValue('26');
            expect(thumb).toHaveValue('26');
            expect(onValueChange).toHaveBeenLastCalledWith(26, expect.objectContaining({ reason: 'input-change' }));
            expect(onValueCommitted).toHaveBeenLastCalledWith(26, expect.objectContaining({ reason: 'input-change' }));

            act(() => thumb.focus());
            await user.keyboard('{ArrowRight}');

            expect(thumb).toHaveValue('27');
            expect(input).toHaveValue('27');
            expect(onValueChange).toHaveBeenLastCalledWith(27, expect.objectContaining({ reason: 'keyboard' }));
        }
    );

    it('renders unique in-range tick labels without adding keyboard stops', () => {
        render(<Slider ticks={[-1, 0, 25, 25, 100, 101, NaN]} thumbAriaLabel='Value' />);

        expect(screen.getByText('0%')).toBeInTheDocument();
        expect(screen.getAllByText('25%')).toHaveLength(1);
        expect(screen.getByText('100%')).toBeInTheDocument();
        expect(screen.queryByText('-1%')).not.toBeInTheDocument();
        expect(screen.queryByText('101%')).not.toBeInTheDocument();
        expect(screen.getAllByRole('slider')).toHaveLength(1);
        expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    });

    it('keeps the value field synchronized with keyboard changes', async () => {
        const user = userEvent.setup();
        render(<Slider defaultValue={25} showValue thumbAriaLabel='Opacity' />);

        await user.tab();
        await user.keyboard('{ArrowRight}');

        expect(screen.getByRole('textbox', { name: 'Opacity value' })).toHaveValue('26');
    });

    it('commits typed values on Enter and snaps to the step relative to min', async () => {
        const user = userEvent.setup();
        const onValueChange = vi.fn();
        const onValueCommitted = vi.fn();
        render(
            <Slider
                min={-1}
                max={1}
                step={0.25}
                defaultValue={0}
                showValue
                thumbAriaLabel='Value'
                valueSuffix=''
                onValueChange={onValueChange}
                onValueCommitted={onValueCommitted}
            />
        );
        const input = screen.getByRole('textbox');

        await user.clear(input);
        await user.type(input, '-0.6{Enter}');

        expect(input).toHaveValue('-0.5');
        expect(screen.getByRole('slider')).toHaveValue('-0.5');
        expect(onValueChange).toHaveBeenCalledWith(-0.5, expect.objectContaining({ reason: 'input-change' }));
        expect(onValueCommitted).toHaveBeenCalledWith(-0.5, expect.objectContaining({ reason: 'input-change' }));
    });

    it('commits on blur and discards edits on Escape', async () => {
        const user = userEvent.setup();
        render(<Slider defaultValue={25} showValue thumbAriaLabel='Value' />);
        const input = screen.getByRole('textbox');

        await user.clear(input);
        await user.type(input, '75');
        await user.tab();
        expect(screen.getByRole('slider')).toHaveValue('75');

        await user.clear(input);
        await user.type(input, '40{Escape}');
        expect(input).toHaveValue('75');
    });

    it.each(['101', '-1', 'abc', ''])('rejects invalid input %j and describes the error', async (draft) => {
        const user = userEvent.setup();
        const onValueChange = vi.fn();
        render(<Slider defaultValue={25} showValue onValueChange={onValueChange} thumbAriaLabel='Value' />);
        const input = screen.getByRole('textbox');

        await user.clear(input);
        if (draft) await user.type(input, draft);
        await user.keyboard('{Enter}');

        expect(input).toHaveAttribute('aria-invalid', 'true');
        expect(input).toHaveAccessibleDescription('Enter a value between 0 and 100. %');
        expect(screen.getByRole('slider')).toHaveValue('25');
        expect(onValueChange).not.toHaveBeenCalled();
    });

    it('respects canceled field changes and does not commit them', async () => {
        const user = userEvent.setup();
        const onValueCommitted = vi.fn();
        render(
            <Slider
                defaultValue={25}
                showValue
                thumbAriaLabel='Value'
                onValueChange={(_, details) => details.cancel()}
                onValueCommitted={onValueCommitted}
            />
        );
        const input = screen.getByRole('textbox');

        await user.clear(input);
        await user.type(input, '75{Enter}');

        expect(input).toHaveValue('25');
        expect(screen.getByRole('slider')).toHaveValue('25');
        expect(onValueCommitted).not.toHaveBeenCalled();
    });

    it('keeps controlled field values aligned with the parent', async () => {
        const user = userEvent.setup();
        const { rerender } = render(<Slider value={25} showValue thumbAriaLabel='Value' />);
        const input = screen.getByRole('textbox');

        await user.clear(input);
        await user.type(input, '75{Enter}');
        expect(input).toHaveValue('25');
        expect(screen.getByRole('slider')).toHaveValue('25');

        rerender(<Slider value={60} showValue thumbAriaLabel='Value' />);
        expect(input).toHaveValue('60');
        expect(screen.getByRole('slider')).toHaveValue('60');
    });

    it('disables the value field together with the thumb', () => {
        render(<Slider disabled showValue thumbAriaLabel='Value' />);

        expect(screen.getByRole('textbox')).toBeDisabled();
        expect(screen.getByRole('slider')).toBeDisabled();
    });
});
