import {
    Checkbox,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
} from '@bloodhoundenterprise/doodleui';
import { useState, type FC } from 'react';
import { COLLECTION_MAP, typedEntries, type EnabledCollections } from '../../utils';

type Props = {
    enabledCollections: EnabledCollections;
    onSelect: (value: string) => void;
};

export const DataCollectedSelect: FC<Props> = ({ enabledCollections, onSelect }) => {
    /*
      Doodle UI's Select is based on Radix UI Select, which does not support
      multi-select. The following behaviors re-implement multi-select:
      
      * Select's `open` state is managed, only closing on outside click, allowing
        multiple selections to be made without close
      * Individual checked states are managed by the parent via props
      * While open, Select's value is set to 'none' via `lastKey` state. A
        non-empty value must be set so that the default SelectValue is not
        triggered.
      * The `lastKey` state is used to manipulate onValueChange behaviors. See
        the comments at lines 57-64 for how this works.
      * Finally, when the user clicks outside the Select, the menu is closed
        and the value is set to match `selected`
      * It's funky, but it works well with not much code
    */
    const [isOpen, setIsOpen] = useState(false);
    const [isOpenable, setIsOpenable] = useState(true);
    const [lastKey, setLastKey] = useState('none');

    const selectedLength = Object.keys(enabledCollections).length;

    const closeSelect = () => {
        setIsOpenable(true);
        setIsOpen(false);
    };

    const openSelect = () => {
        if (isOpenable) {
            setIsOpen(true);
            setIsOpenable(false);
        }
    };

    const toggleOption = (key: string) => {
        onSelect(key);

        // When an option is clicked, Select's value is set to the clicked
        // option's value. `onValueChange` is only called if the value has
        // changed to something new. If the same item is clicked twice in a
        // row, the expected behavior would be that the item was checked and
        // then immediately unchecked. But since the Select's value will not
        // have changed, the `onValueChange` will not trigger until another
        // option is selected. This is why the state is immediately reset to a
        // lastKey value. The actual state is tracked in `enabledCollections`.
        setLastKey('none');
    };

    return (
        <div className='flex flex-col gap-2'>
            <Label htmlFor='data collected'>Data Collected</Label>

            <Select open={isOpen} onOpenChange={openSelect} onValueChange={toggleOption} value={lastKey || ''}>
                <SelectTrigger className='w-32' aria-label='Data Collected Select'>
                    <SelectValue asChild>
                        <span>{selectedLength === 0 ? 'Select' : `${selectedLength} selected`}</span>
                    </SelectValue>
                </SelectTrigger>

                <SelectPortal>
                    <SelectContent onPointerDownOutside={closeSelect}>
                        {typedEntries(COLLECTION_MAP).map(([key, label]) => (
                            <SelectItem key={key} value={key}>
                                <Checkbox size='md' checked={enabledCollections[key] || false} />
                                <span className='ml-2'>{label}</span>
                            </SelectItem>
                        ))}
                    </SelectContent>
                </SelectPortal>
            </Select>
        </div>
    );
};
