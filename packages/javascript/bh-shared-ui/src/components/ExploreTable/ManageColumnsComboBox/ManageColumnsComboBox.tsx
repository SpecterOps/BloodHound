import { Button, Input } from '@bloodhoundenterprise/doodleui';
import { faMinus, faPlus, faRefresh, faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useCombobox, useMultipleSelection } from 'downshift';
import { useMemo, useRef, useState } from 'react';
import { useOnClickOutside } from '../../../hooks';
import { makeStoreMapFromColumnOptions } from '../explore-table-utils';
import ManageColumnsListItem from './ManageColumnsListItem';

export type ManageColumnsComboBoxOption = { id: string; value: string; isPinned?: boolean };

type ManageColumnsComboBoxProps = {
    allColumns: ManageColumnsComboBoxOption[];
    onChange: (items: ManageColumnsComboBoxOption[]) => void;
    selectedColumns: Record<string, boolean>;
};
export const ManageColumnsComboBox = ({
    allColumns,
    onChange = () => {},
    selectedColumns: selectedColumnsProp,
}: ManageColumnsComboBoxProps) => {
    const ref = useRef<HTMLDivElement>(null);

    const [inputValue, setInputValue] = useState('');
    const [isOpen, setIsOpen] = useState(false);

    useOnClickOutside(ref, () => setIsOpen(false));

    const initialColumns = useMemo(
        () => allColumns.filter((item) => selectedColumnsProp[item.id]),
        [allColumns, selectedColumnsProp]
    );
    const pinnedColumns = useMemo(() => allColumns.filter((item) => item.isPinned), [allColumns]);
    const [selectedColumns, setSelectedColumns] = useState<ManageColumnsComboBoxOption[]>(initialColumns);
    const selectedColumnMap = useMemo(() => makeStoreMapFromColumnOptions(selectedColumns), [selectedColumns]);

    const unselectedColumns = useMemo(() => {
        const lowerCasedInputValue = inputValue.toLowerCase();

        return allColumns.filter((column) => {
            const passesFilter = !lowerCasedInputValue || column.value.toLowerCase().includes(lowerCasedInputValue);

            return passesFilter && !column.isPinned && !selectedColumnMap[column.id];
        });
    }, [allColumns, selectedColumnMap, inputValue]);

    const shouldSelectAll = useMemo(() => selectedColumns.length !== allColumns.length, [selectedColumns, allColumns]);

    const { getDropdownProps, removeSelectedItem, addSelectedItem } = useMultipleSelection({
        initialSelectedItems: allColumns.filter((item) => selectedColumnsProp[item.id]),
        selectedItems: selectedColumns,
        onStateChange({ selectedItems: newSelectedColumns, type }) {
            if (type !== useMultipleSelection.stateChangeTypes.DropdownKeyDownBackspace) {
                setSelectedColumns(newSelectedColumns || []);
                onChange(newSelectedColumns || []);
            }
        },
    });

    const { getMenuProps, getInputProps, getItemProps, getComboboxProps } = useCombobox({
        items: unselectedColumns,
        itemToString: (column) => column?.value || '',
        defaultHighlightedIndex: 0, // after selection, highlight the first item.
        selectedItem: null,
        inputValue,
        onStateChange({ inputValue: newInputValue, type, selectedItem: newSelectedItem }) {
            switch (type) {
                case useCombobox.stateChangeTypes.ItemClick:
                    if (newSelectedItem) {
                        setSelectedColumns([...selectedColumns, newSelectedItem]);
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

    const handleResetDefault = () => {
        setSelectedColumns([...pinnedColumns]);
        onChange([...pinnedColumns]);
    };

    const handleSelectAll = () => {
        if (shouldSelectAll) {
            handleResetDefault();
            setSelectedColumns([...allColumns]);
            onChange([...allColumns]);
        } else {
            handleResetDefault();
        }
        return shouldSelectAll;
    };

    const handleManageColumnsClick = () => setIsOpen(true);

    return (
        <>
            <div className='mb-1'>
                <Button
                    onClick={handleManageColumnsClick}
                    className='hover:bg-gray-300 cursor-pointer bg-slate-200 h-8 text-black rounded-full text-sm text-center'>
                    Manage Columns
                </Button>
            </div>

            <div className={`${isOpen ? '' : 'hidden'} absolute z-20 top-3`} ref={ref}>
                <div className='w-[400px] shadow-md border-1 bg-white dark:bg-neutral-dark-5' {...getComboboxProps()}>
                    <div className='flex flex-col gap-1 justify-center'>
                        <div className='flex justify-center items-center relative'>
                            <Input
                                className='border-0 focus:outline-none rounded-none border-black bg-inherit'
                                {...getInputProps(getDropdownProps())}
                            />
                            <FontAwesomeIcon icon={faSearch} className='absolute right-2' />
                        </div>
                    </div>
                    <div className='flex justify-between p-2 border-w-10 border-y border-solid border-neutral-950'>
                        <button className='flex items-center focus:outline-none' onClick={handleSelectAll}>
                            <FontAwesomeIcon icon={shouldSelectAll ? faPlus : faMinus} className='mr-2' />{' '}
                            {shouldSelectAll ? 'Select All' : 'Clear All'}
                        </button>
                        <button className='flex items-center focus:outline-none' onClick={handleResetDefault}>
                            <FontAwesomeIcon icon={faRefresh} className='mr-2' /> Reset Default
                        </button>
                    </div>
                    <ul className={`w-inherit max-h-80 overflow-auto ${!isOpen && 'hidden'}`} {...getMenuProps()}>
                        {isOpen && [
                            ...pinnedColumns.map((column, index) => (
                                <ManageColumnsListItem
                                    isSelected
                                    key={`${column?.id}-${index}`}
                                    item={column}
                                    onClick={removeSelectedItem}
                                    itemProps={getItemProps({ item: column, index })}
                                />
                            )),
                            ...selectedColumns.map((column, index) => {
                                if (!column?.isPinned) {
                                    return (
                                        <ManageColumnsListItem
                                            isSelected
                                            key={`${column?.id}-${index}`}
                                            item={column}
                                            onClick={removeSelectedItem}
                                            itemProps={getItemProps({ item: column, index })}
                                        />
                                    );
                                }
                            }),
                            ...unselectedColumns.map((column, index) => (
                                <ManageColumnsListItem
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
