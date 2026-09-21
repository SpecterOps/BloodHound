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
import { Slider } from './Slider';

expect.extend(matchers);

describe('Slider Tests', () => {
    it('renders a thumb exposing the provided aria-label', () => {
        render(<Slider defaultValue={40} thumbAriaLabel='Horizontal spacing' />);

        expect(screen.getByRole('slider', { name: 'Horizontal spacing' })).toBeInTheDocument();
    });

    it('calls onValueChange when the value is changed via the keyboard', async () => {
        const user = userEvent.setup();
        const onValueChange = vi.fn();

        render(
            <Slider
                defaultValue={50}
                min={0}
                max={100}
                step={1}
                thumbAriaLabel='Value'
                onValueChange={onValueChange}
            />
        );

        const thumb = screen.getByRole('slider', { name: 'Value' });
        thumb.focus();
        await user.keyboard('{ArrowRight}');

        expect(onValueChange).toHaveBeenCalled();
    });

    it('does not call onValueChange when disabled', async () => {
        const user = userEvent.setup();
        const onValueChange = vi.fn();

        render(
            <Slider defaultValue={50} disabled thumbAriaLabel='Value' onValueChange={onValueChange} />
        );

        const thumb = screen.getByRole('slider', { name: 'Value' });
        thumb.focus();
        await user.keyboard('{ArrowRight}');

        expect(onValueChange).not.toHaveBeenCalled();
    });
});
