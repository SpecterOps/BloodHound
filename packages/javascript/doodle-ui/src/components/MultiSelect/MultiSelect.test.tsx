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
import * as React from 'react';
import { afterEach, expect, vi } from 'vitest';
import type { MultiSelectProps } from './MultiSelect';
import { MultiSelect } from './MultiSelect';

expect.extend(matchers);

const options = [
    { value: 'a', label: 'Menu Item A' },
    { value: 'b', label: 'Menu Item B' },
    { value: 'c', label: 'Menu Item C', disabled: true },
];

vi.stubGlobal(
    'ResizeObserver',
    class ResizeObserver {
        observe() {}
        unobserve() {}
        disconnect() {}
    }
);

afterEach(() => {
    cleanup();
});

const renderControlledMultiSelect = (
    props: Omit<MultiSelectProps, 'value' | 'onValueChange'>,
    initialValue: string[] = []
) => {
    const onValueChange = vi.fn();

    const TestMultiSelect = () => {
        const [value, setValue] = React.useState(initialValue);

        const handleValueChange = (nextValue: string[]) => {
            setValue(nextValue);
            onValueChange(nextValue);
        };

        return <MultiSelect {...props} value={value} onValueChange={handleValueChange} />;
    };

    render(<TestMultiSelect />);

    return { onValueChange };
};

describe('MultiSelect', () => {
    it.each([
        { value: [], triggerText: 'All Items' },
        { value: ['a'], triggerText: 'Menu Item A' },
        { value: ['a', 'b'], triggerText: '2 Selected' },
    ])('displays the correct trigger text for the value', ({ value, triggerText }) => {
        render(<MultiSelect options={options} value={value} onValueChange={vi.fn()} placeholder='All Items' />);

        expect(screen.getByRole('button', { name: triggerText })).toBeInTheDocument();
    });

    it('selects and deselects an option while keeping the dropdown open', async () => {
        const user = userEvent.setup();
        const { onValueChange } = renderControlledMultiSelect({
            options,
            placeholder: 'All Items',
        });

        const trigger = screen.getByRole('button', { name: 'All Items' });

        await user.click(trigger);

        expect(screen.getByRole('checkbox', { name: 'Menu Item A' })).toBeInTheDocument();
        expect(screen.getByRole('checkbox', { name: 'Menu Item B' })).toBeInTheDocument();
        expect(screen.getByRole('checkbox', { name: 'Menu Item C' })).toBeDisabled();

        await user.click(screen.getByText('Menu Item A'));

        expect(onValueChange).toHaveBeenLastCalledWith(['a']);
        expect(screen.getByRole('checkbox', { name: 'Menu Item A' })).toBeChecked();
        expect(trigger).toHaveAttribute('aria-expanded', 'true');

        await user.click(screen.getByRole('checkbox', { name: 'Menu Item A' }));

        expect(onValueChange).toHaveBeenLastCalledWith([]);
        expect(screen.getByRole('checkbox', { name: 'Menu Item A' })).not.toBeChecked();
        expect(trigger).toHaveAttribute('aria-expanded', 'true');
    });

    it('does not select a disabled option', async () => {
        const user = userEvent.setup();
        const { onValueChange } = renderControlledMultiSelect({
            options,
            placeholder: 'All Items',
        });

        await user.click(screen.getByRole('button', { name: 'All Items' }));
        await user.click(screen.getByText('Menu Item C'));

        expect(onValueChange).not.toHaveBeenCalled();
        expect(screen.getByRole('checkbox', { name: 'Menu Item C' })).not.toBeChecked();
    });

    it('does not open when disabled', async () => {
        const user = userEvent.setup();

        render(<MultiSelect options={options} value={[]} onValueChange={vi.fn()} placeholder='All Items' disabled />);

        const trigger = screen.getByRole('button', { name: 'All Items' });

        expect(trigger).toBeDisabled();

        await user.click(trigger);

        expect(screen.queryByRole('checkbox', { name: 'Menu Item A' })).not.toBeInTheDocument();
    });

    it('selects and clears all enabled options', async () => {
        const user = userEvent.setup();
        const { onValueChange } = renderControlledMultiSelect(
            {
                options,
                placeholder: 'All Items',
                selectAllLabel: 'All Items',
            },
            ['a']
        );

        await user.click(screen.getByRole('button', { name: 'Menu Item A' }));

        const selectAllCheckbox = screen.getByRole('checkbox', {
            name: 'All Items',
        });

        expect(selectAllCheckbox).toHaveAttribute('aria-checked', 'mixed');

        await user.click(selectAllCheckbox);

        expect(onValueChange).toHaveBeenLastCalledWith(['a', 'b']);
        expect(selectAllCheckbox).toBeChecked();

        await user.click(selectAllCheckbox);

        expect(onValueChange).toHaveBeenLastCalledWith([]);
        expect(selectAllCheckbox).not.toBeChecked();
    });

    it('filters options in any case', async () => {
        const user = userEvent.setup();

        renderControlledMultiSelect({
            options,
            placeholder: 'All Items',
            selectAllLabel: 'All Items',
            isSearchable: true,
            searchPlaceholder: 'Search Items',
            noResultsText: 'No matches',
        });

        await user.click(screen.getByRole('button', { name: 'All Items' }));

        const searchInput = screen.getByRole('textbox', {
            name: 'Search Items',
        });

        await user.type(searchInput, 'mEnU iTeM a');

        expect(screen.getByRole('checkbox', { name: 'Menu Item A' })).toBeInTheDocument();
        expect(screen.queryByRole('checkbox', { name: 'Menu Item B' })).not.toBeInTheDocument();
        expect(screen.queryByRole('checkbox', { name: 'All Items' })).not.toBeInTheDocument();
    });

    it('displays the no results state when no options match', async () => {
        const user = userEvent.setup();

        renderControlledMultiSelect({
            options,
            placeholder: 'All Items',
            selectAllLabel: 'All Items',
            isSearchable: true,
            searchPlaceholder: 'Search Items',
            noResultsText: 'No matches',
        });

        await user.click(screen.getByRole('button', { name: 'All Items' }));

        const searchInput = screen.getByRole('textbox', {
            name: 'Search Items',
        });

        await user.type(searchInput, 'Menu Item F');

        expect(screen.getByText('No matches')).toBeInTheDocument();
        expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
    });

    it('closes with Escape and clears the search before reopening', async () => {
        const user = userEvent.setup();

        renderControlledMultiSelect({
            options,
            placeholder: 'All Items',
            isSearchable: true,
            searchPlaceholder: 'All Items',
        });

        const trigger = screen.getByRole('button', { name: 'All Items' });

        await user.click(trigger);
        await user.type(screen.getByRole('textbox', { name: 'All Items' }), 'a');
        await user.keyboard('{Escape}');

        expect(trigger).toHaveAttribute('aria-expanded', 'false');

        await user.click(trigger);

        expect(screen.getByRole('textbox', { name: 'All Items' })).toHaveValue('');
        expect(screen.getByRole('checkbox', { name: 'Menu Item A' })).toBeInTheDocument();
        expect(screen.getByRole('checkbox', { name: 'Menu Item B' })).toBeInTheDocument();
    });

    it('closes when clicking outside', async () => {
        const user = userEvent.setup();

        renderControlledMultiSelect({
            options,
            placeholder: 'All Items',
        });

        const trigger = screen.getByRole('button', { name: 'All Items' });

        await user.click(trigger);
        expect(trigger).toHaveAttribute('aria-expanded', 'true');

        await user.click(document.body);

        expect(trigger).toHaveAttribute('aria-expanded', 'false');
    });

    it('renders the error state', () => {
        render(<MultiSelect options={options} value={[]} onValueChange={vi.fn()} placeholder='All Items' error />);

        expect(screen.getByRole('button', { name: 'All Items' })).toHaveAttribute('aria-invalid', 'true');
    });

    it('renders the empty state', async () => {
        const user = userEvent.setup();

        renderControlledMultiSelect({
            options: [],
            placeholder: 'All Items',
            emptyText: 'No options available',
        });

        await user.click(screen.getByRole('button', { name: 'All Items' }));

        expect(screen.getByText('No options available')).toBeInTheDocument();
    });

    it('renders the loading state instead of options', async () => {
        const user = userEvent.setup();

        renderControlledMultiSelect({
            options,
            placeholder: 'All Items',
            isLoading: true,
            loadingText: 'Loading Items',
        });

        await user.click(screen.getByRole('button', { name: 'All Items' }));

        expect(screen.getByRole('status', { name: 'Loading Items' })).toBeInTheDocument();
        expect(screen.queryByRole('checkbox', { name: 'Menu Item A' })).not.toBeInTheDocument();
    });
});
