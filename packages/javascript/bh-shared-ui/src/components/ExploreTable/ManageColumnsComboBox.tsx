import { Checkbox, Input } from '@bloodhoundenterprise/doodleui';
import { faMinus, faPlus, faRefresh, faSearch, faThumbTack } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { UseComboboxPropGetters, useCombobox, useMultipleSelection } from 'downshift';
import React, { useEffect, useMemo, useRef } from 'react';

export type ManageColumnsComboBoxOption = { id: string; value: string; isPinned?: boolean };

type ListItemProps = {
    isSelected?: boolean;
    item: ManageColumnsComboBoxOption;
    removeSelectedItem: ReturnType<typeof useMultipleSelection<ManageColumnsComboBoxOption>>['removeSelectedItem'];
    itemProps: ReturnType<UseComboboxPropGetters<ManageColumnsComboBoxOption>['getItemProps']>;
};

const ListItem = ({ isSelected, item, removeSelectedItem, itemProps }: ListItemProps) => (
    <li className='p-2 hover:bg-gray-100 w-full'>
        <button
            className='w-full text-left flex justify-between items-center cursor-default'
            {...itemProps}
            disabled={item?.isPinned}
            onClick={
                isSelected
                    ? (e) => {
                          if (isSelected) {
                              e.stopPropagation();
                              removeSelectedItem(item);
                          }
                      }
                    : itemProps.onClick
            }>
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
    onClose: () => void;
};

export const ManageColumnsComboBox = ({
    allItems,
    onChange = () => {},
    onClose,
    visibleColumns,
}: ManageColumnsComboBoxProps) => {
    const [inputValue, setInputValue] = React.useState('');
    const pinnedItems = useMemo(() => allItems.filter((item) => item.isPinned), [allItems]);
    const initialVisibleColumns = useMemo(
        () => allItems.filter((item) => visibleColumns[item.id]),
        [allItems, visibleColumns]
    );

    console.log({ visibleColumns });

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
        const listener = (e: Event) => {
            if (!ref.current || ref.current?.contains(e.target)) {
                return;
            }

            onClose();
        };
        document.addEventListener('mousedown', listener);
        document.addEventListener('touchstart', listener);

        return () => {
            document.removeEventListener('mousedown', listener);
            document.removeEventListener('touchstart', listener);
        };
    }, [onClose]);

    useEffect(() => {
        const selectedItems = allItems.filter((item) => visibleColumns[item.id]);
        setSelectedItems(selectedItems);
    }, [visibleColumns, allItems]);

    const { getDropdownProps, removeSelectedItem } = useMultipleSelection({
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

    const isOpen = true;
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
        console.log('resetting');
        setSelectedItems([...pinnedItems]);
    };

    const handleSelectAll = () => {
        if (shouldSelectAll) {
            console.log('selecting all');
            setSelectedItems([...allItems]);
        } else {
            console.log('reset');
            handleResetDefault();
        }
        return shouldSelectAll;
    };

    // console.log('\n////////////////');
    // console.log({ unselectedItems });
    // console.log({ selectedItems });
    // console.log({ pinnedItems });

    return (
        <div className='w-[400px] shadow-md border-1 bg-white' ref={ref}>
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
                        <ListItem
                            isSelected
                            key={`${item?.id}${index}`}
                            item={item}
                            removeSelectedItem={removeSelectedItem}
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
                                    removeSelectedItem={removeSelectedItem}
                                    itemProps={getItemProps({ item, index })}
                                />
                            );
                        }
                    }),
                    ...unselectedItems.map((item, index) => (
                        <ListItem
                            key={`${item?.id}${index}`}
                            item={item}
                            removeSelectedItem={removeSelectedItem}
                            itemProps={getItemProps({ item, index })}
                        />
                    )),
                ]}
            </ul>
        </div>
    );
};
