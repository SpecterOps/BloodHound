import { Button, Checkbox, Input } from '@bloodhoundenterprise/doodleui';
import { faMinus, faPlus, faRefresh, faSearch, faThumbTack } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { UseComboboxPropGetters, useCombobox, useMultipleSelection } from 'downshift';
import React, { useEffect, useMemo, useRef } from 'react';

export type ManageColumnsComboBoxOption = { id: string; value: string; isPinned?: boolean };

type ListItemProps = {
    isSelected?: boolean;
    item: ManageColumnsComboBoxOption;
    onClick:
        | ReturnType<typeof useMultipleSelection<ManageColumnsComboBoxOption>>['removeSelectedItem']
        | ReturnType<typeof useMultipleSelection<ManageColumnsComboBoxOption>>['addSelectedItem'];
    itemProps: ReturnType<UseComboboxPropGetters<ManageColumnsComboBoxOption>['getItemProps']>;
};

const ListItem = ({ isSelected, item, onClick, itemProps }: ListItemProps) => (
    <li
        className='p-2 hover:bg-gray-100 w-full'
        {...itemProps}
        disabled={item?.isPinned}
        onClick={(e) => {
            e.stopPropagation();
            onClick(item);
        }}>
        <button className='w-full text-left flex justify-between items-center cursor-default'>
            <div>
                <Checkbox className={`mr-2 ${isSelected ? `&:*['bg-blue-800']` : ''}`} checked={isSelected} />
                <span>{item.value}</span>
            </div>
            {item.isPinned && (
                <FontAwesomeIcon
                    fill='white'
                    stroke=''
                    className='justify-self-end stroke-cyan-300'
                    icon={faThumbTack}
                />
            )}
        </button>
    </li>
);

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
    const [inputValue, setInputValue] = React.useState('');
    const pinnedItems = useMemo(() => allItems.filter((item) => item.isPinned), [allItems]);
    const initialVisibleColumns = useMemo(
        () => allItems.filter((item) => visibleColumns[item.id]),
        [allItems, visibleColumns]
    );

    const [selectedItems, setSelectedItems] = React.useState(initialVisibleColumns);
    const unselectedItems = React.useMemo(() => {
        const lowerCasedInputValue = inputValue.toLowerCase();

        return allItems.filter(
            (item) => !selectedItems.includes(item) && item.value.toLowerCase().includes(lowerCasedInputValue)
        );
    }, [allItems, selectedItems, inputValue]);

    const shouldSelectAll = useMemo(() => selectedItems.length !== allItems.length, [selectedItems, allItems]);

    const ref = useRef<HTMLDivElement>();

    useEffect(() => {
        const selectedItems = allItems.filter((item) => visibleColumns[item.id]);
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

    const { getMenuProps, getInputProps, getItemProps, getToggleButtonProps, isOpen } = useCombobox({
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
    };

    const handleSelectAll = () => {
        if (shouldSelectAll) {
            setSelectedItems([...allItems]);
        } else {
            handleResetDefault();
        }
        return shouldSelectAll;
    };

    return (
        <>
            <div className='mb-1'>
                <Button
                    className='hover:bg-gray-300 cursor-pointer bg-slate-200 h-8 text-black rounded-full text-sm text-center'
                    {...getToggleButtonProps()}>
                    Manage Columns
                </Button>
            </div>
            {isOpen && (
                <div className='absolute z-20 top-3'>
                    <div className='w-[400px] shadow-md border-1 bg-white' ref={() => ref}>
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
                        <ul
                            className={`w-inherit mt-1 max-h-80 overflow-auto ${!isOpen && 'hidden'}`}
                            {...getMenuProps()}>
                            {isOpen && [
                                ...pinnedItems.map((item, index) => (
                                    <ListItem
                                        isSelected
                                        key={`${item?.id}${index}`}
                                        item={item}
                                        onClick={removeSelectedItem}
                                        itemProps={getItemProps({ item, index })}
                                    />
                                )),
                                ...selectedItems.map((item, index) => {
                                    if (!item?.isPinned) {
                                        return (
                                            <ListItem
                                                isSelected
                                                key={`${item?.id}${index}`}
                                                item={item}
                                                onClick={removeSelectedItem}
                                                itemProps={getItemProps({ item, index })}
                                            />
                                        );
                                    }
                                }),
                                ...unselectedItems.map((item, index) => (
                                    <ListItem
                                        key={`${item?.id}${index}`}
                                        item={item}
                                        onClick={addSelectedItem}
                                        itemProps={getItemProps({ item, index })}
                                    />
                                )),
                            ]}
                        </ul>
                    </div>
                </div>
            )}
        </>
    );
};
