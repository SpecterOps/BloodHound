import { Button, Input } from '@bloodhoundenterprise/doodleui';
import { faMinus, faPlus, faRefresh, faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useCombobox, useMultipleSelection } from 'downshift';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useOnClickOutside } from '../../../hooks';
import ManageColumnsListItem from './ManageColumnsListItem';

export type ManageColumnsComboBoxOption = { id: string; value: string; isPinned?: boolean };

type ManageColumnsComboBoxProps = {
    allItems: ManageColumnsComboBoxOption[];
    onChange: (items: ManageColumnsComboBoxOption[]) => void;
    visibleColumns: Record<string, boolean>;
};

export const ManageColumnsComboBox = ({
    allItems,
    onChange = () => {},
    visibleColumns,
}: ManageColumnsComboBoxProps) => {
    const ref = useRef<HTMLDivElement>(null);

    const [inputValue, setInputValue] = useState('');
    const [isOpen, setIsOpen] = useState(false);

    useOnClickOutside(ref, () => setIsOpen(false));

    const pinnedItems = useMemo(() => allItems.filter((item) => item.isPinned), [allItems]);
    const [selectedItems, setSelectedItems] = useState<ManageColumnsComboBoxOption[]>([]);
    const unselectedItems = useMemo(() => {
        const lowerCasedInputValue = inputValue.toLowerCase();

        return allItems.filter(
            (item) =>
                !item.isPinned &&
                !selectedItems.includes(item) &&
                item.value.toLowerCase().includes(lowerCasedInputValue)
        );
    }, [allItems, selectedItems, inputValue]);

    const shouldSelectAll = useMemo(() => selectedItems.length !== allItems.length, [selectedItems, allItems]);

    useEffect(() => {
        const selectedItems = allItems.filter((item) => visibleColumns[item.id] && !item.isPinned);
        setSelectedItems(selectedItems);
    }, [visibleColumns, allItems]);

    const { getDropdownProps, removeSelectedItem, addSelectedItem } = useMultipleSelection({
        initialSelectedItems: allItems.filter((item) => visibleColumns[item.id]),
        selectedItems,
        onStateChange({ selectedItems: newSelectedItems, type }) {
            onChange(newSelectedItems || []);

            switch (type) {
                case useMultipleSelection.stateChangeTypes.SelectedItemKeyDownBackspace:
                case useMultipleSelection.stateChangeTypes.SelectedItemKeyDownDelete:
                case useMultipleSelection.stateChangeTypes.DropdownKeyDownBackspace:
                case useMultipleSelection.stateChangeTypes.FunctionRemoveSelectedItem:
                    setSelectedItems(newSelectedItems || []);
                    break;
                default:
                    break;
            }
        },
    });

    const { getMenuProps, getInputProps, getItemProps } = useCombobox({
        items: unselectedItems,
        itemToString: (item) => item?.value || '',
        defaultHighlightedIndex: 0, // after selection, highlight the first item.
        selectedItem: null,
        inputValue,
        onStateChange({ inputValue: newInputValue, type, selectedItem: newSelectedItem }) {
            switch (type) {
                case useCombobox.stateChangeTypes.InputKeyDownEnter:
                case useCombobox.stateChangeTypes.ItemClick:
                case useCombobox.stateChangeTypes.InputBlur:
                    if (newSelectedItem) {
                        setSelectedItems([...selectedItems, newSelectedItem]);
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
        setSelectedItems([...pinnedItems]);
        onChange([...pinnedItems]);
    };

    const handleSelectAll = () => {
        if (shouldSelectAll) {
            handleResetDefault();
            setSelectedItems([...allItems]);
            onChange([...allItems]);
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
                <div className='w-[400px] shadow-md border-1 bg-white'>
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
                    <ul className={`w-inherit mt-1 max-h-80 overflow-auto ${!isOpen && 'hidden'}`} {...getMenuProps()}>
                        {isOpen && [
                            ...pinnedItems.map((item, index) => (
                                <ManageColumnsListItem
                                    isSelected
                                    key={item?.id}
                                    item={item}
                                    onClick={removeSelectedItem}
                                    itemProps={getItemProps({ item, index })}
                                />
                            )),
                            ...selectedItems.map((item, index) => {
                                if (!item?.isPinned) {
                                    return (
                                        <ManageColumnsListItem
                                            isSelected
                                            key={item?.id}
                                            item={item}
                                            onClick={removeSelectedItem}
                                            itemProps={getItemProps({ item, index })}
                                        />
                                    );
                                }
                            }),
                            ...unselectedItems.map((item, index) => (
                                <ManageColumnsListItem
                                    key={item?.id}
                                    item={item}
                                    onClick={addSelectedItem}
                                    itemProps={getItemProps({ item, index })}
                                />
                            )),
                        ]}
                    </ul>
                </div>
            </div>
        </>
    );
};
