import { Button, Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { FC, useState } from 'react';
import { AppIcon } from '../../../components';
import TagToZoneDialog from './TagToZoneDialog';

type TagToZoneLabelProps = {
    selectedQuery: any;
};

const TagToZoneLabel: FC<TagToZoneLabelProps> = (props) => {
    const { selectedQuery } = props;
    const listItemStyles = 'px-2 py-3 cursor-pointer hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-4';

    const [tagToZoneOpen, setTagToZoneOpen] = useState(false);
    //Tag to Zone or Label
    const [isLabel, setIsLabel] = useState(false);

    const handleSetOpen = (isOpen: boolean) => {
        setTagToZoneOpen(isOpen);
    };

    const tagToZone = () => {
        setIsLabel(false);
        setTagToZoneOpen(true);
    };

    const tagToLabel = () => {
        setIsLabel(true);
        setTagToZoneOpen(true);
    };

    return (
        <>
            <Popover>
                <PopoverTrigger disabled={!selectedQuery}>
                    <Button variant='secondary' asChild size='small'>
                        <div>
                            <span className='mr-2 text-base'>Tag</span>
                            <AppIcon.CaretDown size={10} />
                        </div>
                    </Button>
                </PopoverTrigger>
                <PopoverContent className='p-0 w-28'>
                    <div className={listItemStyles} onClick={tagToZone}>
                        Zone
                    </div>
                    <div className={listItemStyles} onClick={tagToLabel}>
                        Label
                    </div>
                </PopoverContent>
            </Popover>
            <TagToZoneDialog
                dialogOpen={tagToZoneOpen}
                setDialogOpen={handleSetOpen}
                isLabel={isLabel}
                selectedQuery={selectedQuery}
            />
        </>
    );
};

export default TagToZoneLabel;
