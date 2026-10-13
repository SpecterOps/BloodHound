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
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, vi } from 'vitest';
import { Select, SelectTrigger, SelectValue } from './Select';

expect.extend(matchers);
afterEach(cleanup);

describe('SelectTrigger', () => {
    it('disables the trigger while loading and restores it when loading finishes', async () => {
        const user = userEvent.setup();
        const onOpenChange = vi.fn();
        const renderSelect = (isLoading: boolean) => (
            <Select onOpenChange={onOpenChange}>
                <SelectTrigger aria-label='Schedule' isLoading={isLoading} disabled={false}>
                    <SelectValue placeholder='Select schedule' />
                </SelectTrigger>
            </Select>
        );
        const { rerender } = render(renderSelect(true));
        const trigger = screen.getByRole('combobox', { name: 'Schedule' });

        expect(trigger).toBeDisabled();
        expect(trigger).toHaveAttribute('aria-busy', 'true');
        expect(trigger.querySelector('[data-icon="spinner"]')).toBeInTheDocument();
        expect(trigger.querySelector('[data-icon="chevron-down"]')).not.toBeInTheDocument();
        // jsdom does not implement pointer capture.
        Object.defineProperty(trigger, 'hasPointerCapture', { value: () => false });
        await user.click(trigger);
        expect(onOpenChange).not.toHaveBeenCalled();

        rerender(renderSelect(false));

        expect(trigger).toBeEnabled();
        expect(trigger).toHaveAttribute('aria-busy', 'false');
        expect(trigger.querySelector('[data-icon="spinner"]')).not.toBeInTheDocument();
        expect(trigger.querySelector('[data-icon="chevron-down"]')).toBeInTheDocument();
    });

    it('keeps an explicitly disabled trigger disabled when loading finishes', () => {
        const renderSelect = (isLoading: boolean) => (
            <Select>
                <SelectTrigger aria-label='Schedule' isLoading={isLoading} disabled>
                    <SelectValue placeholder='Select schedule' />
                </SelectTrigger>
            </Select>
        );
        const { rerender } = render(renderSelect(true));

        rerender(renderSelect(false));

        expect(screen.getByRole('combobox', { name: 'Schedule' })).toBeDisabled();
    });
});
