import { Button, Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { FC, useState } from 'react';
import { AppIcon } from '../../../../components';
import { QueryLineItem } from '../../../../types';
import TagToZoneLabelDialog from './TagToZoneLabelDialog';

type TagToZoneLabelProps = {
    selectedQuery: QueryLineItem | undefined;
    cypherQuery: string;
};

const TagToZoneLabel: FC<TagToZoneLabelProps> = (props) => {
    const { selectedQuery, cypherQuery } = props;

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
                <PopoverTrigger disabled={!selectedQuery && !cypherQuery}>
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
            <TagToZoneLabelDialog
                dialogOpen={tagToZoneOpen}
                setDialogOpen={handleSetOpen}
                isLabel={isLabel}
                selectedQuery={selectedQuery}
                cypherQuery={cypherQuery}
            />
        </>
    );
};

export default TagToZoneLabel;
