import { Button, Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { FC, useState } from 'react';
import { AppIcon } from '../../../components';
import TagToZoneDialog from './TagToZoneDialog';
const TagToZoneLabel: FC = () => {
    const listItemStyles = 'px-2 py-3 cursor-pointer hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-4';

    const [tagToZoneOpen, setTagToZoneOpen] = useState(false);

    const handleSetOpen = (isOpen: boolean) => {
        setTagToZoneOpen(isOpen);
    };

    return (
        <>
            <Popover>
                <PopoverTrigger>
                    <Button variant='secondary' asChild size='small'>
                        <div>
                            <span className='mr-2 text-base'>Tag</span>
                            <AppIcon.CaretDown size={10} />
                        </div>
                    </Button>
                </PopoverTrigger>
                <PopoverContent className='p-0 w-28'>
                    <div className={listItemStyles} onClick={() => setTagToZoneOpen(true)}>
                        Zone
                    </div>
                    <div className={listItemStyles} onClick={() => console.log('clicked')}>
                        Label
                    </div>
                </PopoverContent>
            </Popover>
            <TagToZoneDialog dialogOpen={tagToZoneOpen} setDialogOpen={handleSetOpen} />
        </>
    );
};

export default TagToZoneLabel;
