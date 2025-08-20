import {
    Checkbox,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
} from '@bloodhoundenterprise/doodleui';
import { Minus } from 'lucide-react';
import { useState, type FC } from 'react';
import type { Control, FieldValues } from 'react-hook-form';
import { COLLECTION_MAP } from '../../utils';

const collections = Array.from(COLLECTION_MAP);
const minusIcon = <Minus className='h-full w-full' absoluteStrokeWidth={true} strokeWidth={3} />;

export const DataCollectedSelect: FC<{ control: Control<FieldValues, any, FieldValues> }> = ({ control }) => {
    const [selected, setSelected] = useState<Record<string, boolean>>({});

    /*
      Doodle UI's Select is based on Radix UI Select, which does not support
      multi-select. The following behaviors re-implement multi-select:
      
      * Select `open` state is managed, only closing on outside click, allowing
        multiple selections to be made without close
      * While open, Select's value is set to '<key>' via `lastKey` state. A
        non-empty value must be set so that the default SelectValue is not
        triggered.
      * Individual checkbox states are actually managed via `selected` state
      * The `lastKey` state is used to manipulate onValueChange behaviors. See
        the comments at lines 61-68 for how this works.
      * Finally, when the user clicks outside the Select, the menu is closed
        and the value is set to match `selected`
      * It's janky, but it works
      * This is why you don't roll your own component library
    */
    const [isOpen, setIsOpen] = useState(false);
    const [isOpenable, setIsOpenable] = useState(true);
    const [lastKey, setLastKey] = useState('<key>');

    const selectedLength = Object.keys(selected).length;

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

    /**
     * Options have a 3 state toggle:
     * `<key>=true` => `<key>=false` => <key> removed =>
     */
    const toggleOption = (key: string) => {
        if (selected[key] === true) {
            selected[key] = false;
        } else if (selected[key] === false) {
            delete selected[key];
        } else if (selected[key] === undefined) {
            selected[key] = true;
        }

        setSelected({ ...selected });

        // When an option is clicked, Select's value is set to the clicked
        // option's value. `onValueChange` is only called if the value has
        // changed to something new. If the same item is clicked twice in a
        // row, the expected behavior would be that the item was checked and
        // then immediately unchecked. But since the Select's value will not
        // have changed, the `onValueChange` will not trigger until another
        // option is selected. This is why the state is immediately reset to a
        // lastKey value. The actual state is tracked in the `selected` state.
        setLastKey('<key>');
    };

    console.log(selected);

    return (
        <FormField
            control={control}
            name='data_collected'
            render={({ field }) => (
                <FormItem>
                    <FormLabel>Data Collected</FormLabel>

                    <FormControl>
                        <Select open={isOpen} onOpenChange={openSelect} onValueChange={toggleOption} value={lastKey}>
                            <SelectTrigger className='w-32'>
                                <SelectValue asChild>
                                    <span>{selectedLength === 0 ? 'Select' : `${selectedLength} selected`}</span>
                                </SelectValue>
                            </SelectTrigger>

                            <SelectPortal>
                                <SelectContent onPointerDownOutside={closeSelect}>
                                    {collections.map(([key, label]) => (
                                        <SelectItem key={key} value={key}>
                                            {selected[key] === true && <Checkbox size='md' checked />}
                                            {selected[key] === undefined && <Checkbox size='md' />}
                                            {selected[key] === false && <Checkbox icon={minusIcon} size='md' checked />}
                                            <span className='ml-2'>{label}</span>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </SelectPortal>
                        </Select>
                    </FormControl>
                </FormItem>
            )}
        />
    );
};
