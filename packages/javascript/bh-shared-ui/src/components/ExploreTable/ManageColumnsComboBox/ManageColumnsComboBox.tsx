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
import { Button, Input } from '@bloodhoundenterprise/doodleui';
import { faArrowsLeftRight, faMinus, faPlus, faRefresh, faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useCombobox, useMultipleSelection } from 'downshift';
import { ReactNode, useMemo, useRef, useState } from 'react';
import { useOnClickOutside } from '../../../hooks';
import { defaultColumns, makeStoreMapFromColumnOptions } from '../explore-table-utils';
import ManageColumnsListItem from './ManageColumnsListItem';

export type ManageColumnsComboBoxOption = { id: string; value: string; isPinned?: boolean };

type ManageColumnsComboBoxProps = {
    allColumns: ManageColumnsComboBoxOption[];
    disabled?: boolean;
    onChange: (items: ManageColumnsComboBoxOption[]) => void;
    selectedColumns: Record<string, boolean>;
    onResetColumnSize?: () => void;
};
export const ManageColumnsComboBox = ({
    allColumns,
    onChange = () => {},
    disabled,
    selectedColumns: selectedColumnsProp,
    onResetColumnSize,
}: ManageColumnsComboBoxProps) => {
    const ref = useRef<HTMLDivElement>(null);

    const [inputValue, setInputValue] = useState('');
    const [isOpen, setIsOpen] = useState(false);

    useOnClickOutside(ref, () => setIsOpen(false));

    const selectedColumns = useMemo(
        () => allColumns.filter((item) => selectedColumnsProp[item.id]),
        [allColumns, selectedColumnsProp]
    );

    const pinnedColumns = useMemo(() => allColumns.filter((item) => item.isPinned), [allColumns]);
    const initialColumns = useMemo(() => allColumns.filter((item) => defaultColumns[item.id]), [allColumns]);
    const selectedColumnMap = useMemo(() => makeStoreMapFromColumnOptions(selectedColumns), [selectedColumns]);

    const unselectedColumns = useMemo(() => {
        const lowerCasedInputValue = inputValue.toLowerCase();

        return allColumns.filter((column) => {
            const passesFilter = !lowerCasedInputValue || column.value.toLowerCase().includes(lowerCasedInputValue);

            return passesFilter && !column.isPinned && !selectedColumnMap[column.id];
        });
    }, [allColumns, selectedColumnMap, inputValue]);

    const handleResetDefault = () => {
        onChange([...initialColumns]);
    };

    const shouldSelectAll = useMemo(() => selectedColumns.length !== allColumns.length, [selectedColumns, allColumns]);

    const { getDropdownProps, removeSelectedItem, addSelectedItem } = useMultipleSelection({
        initialSelectedItems: allColumns.filter((item) => selectedColumnsProp[item.id]),
        selectedItems: selectedColumns,
        onStateChange({ selectedItems: newSelectedColumns, type }) {
            if (type !== useMultipleSelection.stateChangeTypes.DropdownKeyDownBackspace && newSelectedColumns?.length) {
                onChange(newSelectedColumns || []);
            } else {
                handleResetDefault();
            }
        },
    });

    const { getMenuProps, getInputProps, getItemProps } = useCombobox({
        items: unselectedColumns,
        itemToString: (column) => column?.value || '',
        defaultHighlightedIndex: 0, // after selection, highlight the first item.
        selectedItem: null,
        inputValue,
        onStateChange({ inputValue: newInputValue, type, selectedItem: newSelectedItem }) {
            switch (type) {
                case useCombobox.stateChangeTypes.ItemClick:
                    if (newSelectedItem) {
                        setInputValue('');
                    }
                    break;

                case useCombobox.stateChangeTypes.InputChange:
                    setInputValue(newInputValue || '');

                    break;
                default:
                    break;
            }
        },
    });

    const handleSelectAll = () => {
        if (shouldSelectAll) {
            handleResetDefault();
            onChange([...allColumns]);
        } else {
            handleResetDefault();
        }
        return shouldSelectAll;
    };

    const handleManageColumnsClick = () => {
        setIsOpen(true);
    };

    return (
        <>
            <div className='mb-1'>
                <Button
                    disabled={disabled}
                    onClick={handleManageColumnsClick}
                    className='hover:bg-gray-300 cursor-pointer bg-slate-200 h-8 text-black rounded-full text-sm text-center'>
                    Columns
                </Button>
            </div>

            <div className={`${isOpen ? '' : 'hidden'} absolute z-20 top-16`} ref={ref}>
                <div className='w-[420px] shadow-md border-1 bg-white dark:bg-neutral-dark-5 rounded-md'>
                    <div className='flex flex-col gap-1 justify-center'>
                        <div className='flex justify-center items-center relative'>
                            <Input
                                className='border-0 focus:outline-none rounded-none border-black bg-inherit'
                                aria-label='Filter columns'
                                {...getInputProps(getDropdownProps())}
                            />
                            <FontAwesomeIcon icon={faSearch} className='absolute right-2' />
                        </div>
                    </div>
                    <div className='flex justify-between p-2 border-w-10 border-y border-solid border-neutral-950'>
                        <button className='flex items-center' onClick={handleSelectAll}>
                            <FontAwesomeIcon icon={shouldSelectAll ? faPlus : faMinus} className='mr-2' />{' '}
                            {shouldSelectAll ? 'Select All' : 'Clear All'}
                        </button>
                        <button onClick={onResetColumnSize}>
                            <FontAwesomeIcon icon={faArrowsLeftRight} className='mr-2' />
                            Reset Size
                        </button>
                        <button className='flex items-center' onClick={handleResetDefault}>
                            <FontAwesomeIcon icon={faRefresh} className='mr-2' /> Reset Default
                        </button>
                    </div>
                    <ul className={`w-inherit max-h-60 overflow-auto ${!isOpen && 'hidden'}`} {...getMenuProps()}>
                        {isOpen && [
                            ...pinnedColumns.map((column, index) => {
                                const isSelected = selectedColumnMap[column.id];

                                return (
                                    <ManageColumnsListItem
                                        isSelected={!!selectedColumnMap[column.id]}
                                        key={`${column?.id}-${index}`}
                                        item={column}
                                        onClick={isSelected ? removeSelectedItem : addSelectedItem}
                                        itemProps={getItemProps({ item: column, index })}
                                    />
                                );
                            }),
                            ...selectedColumns.reduce((acc, column, index) => {
                                if (!column?.isPinned) {
                                    acc.push(
                                        <ManageColumnsListItem
                                            isSelected
                                            key={`${column?.id}-${index}`}
                                            item={column}
                                            onClick={removeSelectedItem}
                                            itemProps={getItemProps({ item: column, index })}
                                        />
                                    );
                                }

                                return acc;
                            }, [] as ReactNode[]),
                            ...unselectedColumns.map((column, index) => (
                                <ManageColumnsListItem
                                    isSelected={false}
                                    key={`${column?.id}-${index}`}
                                    item={column}
                                    onClick={addSelectedItem}
                                    itemProps={getItemProps({ item: column, index })}
                                />
                            )),
                        ]}
                    </ul>
                </div>
            </div>
        </>
    );
};
