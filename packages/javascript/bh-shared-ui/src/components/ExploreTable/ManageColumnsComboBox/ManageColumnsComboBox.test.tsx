// Copyright 2025 Specter Ops, Inc.
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

import userEvent from '@testing-library/user-event';
import { render, screen } from '../../../test-utils';
import { DEFAULT_PINNED_COLUMN_KEYS } from '../explore-table-utils';
import { ManageColumnsComboBox, ManageColumnsComboBoxOption } from './ManageColumnsComboBox';

// Override the global FontAwesomeIcon mock so that onClick is forwarded to the rendered span.
// This is required to test pin-icon clicks in ManageColumnsListItem.
vi.mock('@fortawesome/react-fontawesome', () => ({
    FontAwesomeIcon: vi.fn(({ icon, onClick }: any) => {
        const iconName = typeof icon === 'string' ? icon : icon.iconName;
        // eslint-disable-next-line
        return <span onClick={onClick}>{iconName}</span>;
    }),
}));

const allColumns: ManageColumnsComboBoxOption[] = [
    { id: 'kind', value: 'Kind', isPinned: true },
    { id: 'label', value: 'Label', isPinned: true },
    { id: 'objectId', value: 'Object ID' },
    { id: 'isTierZero', value: 'Is Tier Zero' },
    { id: 'custom1', value: 'Custom Column 1' },
];

// kind, label, objectId, isTierZero are in defaultColumns
const defaultSelectedColumns: Record<string, boolean> = {
    kind: true,
    label: true,
    objectId: true,
    isTierZero: true,
};

const setup = (overrides: Partial<Parameters<typeof ManageColumnsComboBox>[0]> = {}) => {
    const onChange = vi.fn();
    const onChangePinnedColumns = vi.fn();
    const onResetColumnSize = vi.fn();
    const user = userEvent.setup();

    render(
        <ManageColumnsComboBox
            allColumns={allColumns}
            selectedColumns={defaultSelectedColumns}
            onChange={onChange}
            onChangePinnedColumns={onChangePinnedColumns}
            onResetColumnSize={onResetColumnSize}
            {...overrides}
        />
    );

    return { onChange, onChangePinnedColumns, onResetColumnSize, user };
};

describe('ManageColumnsComboBox', () => {
    it('renders the Columns trigger button', () => {
        setup();
        expect(screen.getByRole('button', { name: /columns/i })).toBeInTheDocument();
    });

    it('does not show the dropdown by default', () => {
        setup();
        // jsdom does not process Tailwind CSS; check for the 'hidden' class on the container div directly
        const dropdownContainer = screen.getByLabelText('Filter columns').closest('.absolute');
        expect(dropdownContainer).toHaveClass('hidden');
    });

    it('disables the Columns button when the disabled prop is true', () => {
        setup({ disabled: true });
        expect(screen.getByRole('button', { name: /columns/i })).toBeDisabled();
    });

    it('opens the dropdown when the Columns button is clicked', async () => {
        const { user } = setup();
        await user.click(screen.getByRole('button', { name: /columns/i }));
        expect(screen.getByLabelText('Filter columns')).toBeVisible();
    });

    it('shows "Select All" when not all columns are selected', async () => {
        const { user } = setup({ selectedColumns: { kind: true } });
        await user.click(screen.getByRole('button', { name: /columns/i }));
        expect(screen.getByRole('button', { name: /select all/i })).toBeInTheDocument();
    });

    it('shows "Clear All" when all columns are selected', async () => {
        const { user } = setup({
            selectedColumns: Object.fromEntries(allColumns.map((col) => [col.id, true])),
        });
        await user.click(screen.getByRole('button', { name: /columns/i }));
        expect(screen.getByRole('button', { name: /clear all/i })).toBeInTheDocument();
    });

    it('calls onChange with all columns when "Select All" is clicked', async () => {
        const { user, onChange } = setup({ selectedColumns: { kind: true } });
        await user.click(screen.getByRole('button', { name: /columns/i }));
        await user.click(screen.getByRole('button', { name: /select all/i }));
        expect(onChange).toHaveBeenCalledWith(allColumns);
    });

    it('calls onResetColumnSize when "Reset Size" is clicked', async () => {
        const { user, onResetColumnSize } = setup();
        await user.click(screen.getByRole('button', { name: /columns/i }));
        await user.click(screen.getByRole('button', { name: /reset size/i }));
        expect(onResetColumnSize).toHaveBeenCalledTimes(1);
    });

    it('calls onChangePinnedColumns with DEFAULT_PINNED_COLUMN_KEYS when "Reset Default" is clicked', async () => {
        const { user, onChangePinnedColumns } = setup();
        await user.click(screen.getByRole('button', { name: /columns/i }));
        await user.click(screen.getByRole('button', { name: /reset default/i }));
        expect(onChangePinnedColumns).toHaveBeenCalledWith([...DEFAULT_PINNED_COLUMN_KEYS]);
    });

    it('calls onChange with initial default columns when "Reset Default" is clicked', async () => {
        const { user, onChange } = setup();
        await user.click(screen.getByRole('button', { name: /columns/i }));
        await user.click(screen.getByRole('button', { name: /reset default/i }));
        // initialColumns = allColumns filtered by defaultColumns (kind, label, objectId, isTierZero)
        const expectedInitialColumns = allColumns.filter((col) =>
            ['kind', 'label', 'objectId', 'isTierZero'].includes(col.id)
        );
        expect(onChange).toHaveBeenCalledWith(expectedInitialColumns);
    });

    it('displays pinned columns and unselected columns in the dropdown', async () => {
        const { user } = setup({ selectedColumns: {} });
        await user.click(screen.getByRole('button', { name: /columns/i }));
        // Pinned columns are always shown first
        expect(screen.getByText('Kind')).toBeInTheDocument();
        expect(screen.getByText('Label')).toBeInTheDocument();
        // Unselected non-pinned columns are also listed
        expect(screen.getByText('Object ID')).toBeInTheDocument();
        expect(screen.getByText('Custom Column 1')).toBeInTheDocument();
    });

    it('filters unselected columns when typing in the filter input', async () => {
        const { user } = setup({ selectedColumns: {} });
        await user.click(screen.getByRole('button', { name: /columns/i }));
        await user.type(screen.getByLabelText('Filter columns'), 'custom');
        expect(screen.getByText('Custom Column 1')).toBeInTheDocument();
        expect(screen.queryByText('Object ID')).not.toBeInTheDocument();
        expect(screen.queryByText('Is Tier Zero')).not.toBeInTheDocument();
    });

    it('calls onChangePinnedColumns to unpin a column when its pin icon is clicked', async () => {
        const { user, onChangePinnedColumns } = setup();
        await user.click(screen.getByRole('button', { name: /columns/i }));
        // 'Kind' is pinned; clicking its thumbtack should remove it from the pinned list
        const thumbtackIcons = screen.getAllByText('thumbtack');
        await user.click(thumbtackIcons[0]);
        // 'kind' was the first pinned column; clicking should produce a list without 'kind'
        expect(onChangePinnedColumns).toHaveBeenCalledWith(
            allColumns.filter((col) => col.isPinned && col.id !== 'kind').map((col) => col.id)
        );
    });

    it('closes the dropdown when clicking outside', async () => {
        const { user } = setup();
        await user.click(screen.getByRole('button', { name: /columns/i }));
        const dropdownContainer = screen.getByLabelText('Filter columns').closest('.absolute');
        expect(dropdownContainer).not.toHaveClass('hidden');
        await user.click(document.body);
        expect(dropdownContainer).toHaveClass('hidden');
    });
});
