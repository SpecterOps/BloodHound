import { Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { FC, MouseEvent } from 'react';
import { AppIcon } from '../../../components';
// import { VerticalEllipsis } from '../AppIcon/Icons';
interface SaveQueryActionMenuProps {
    saveAs: () => void;
}

const SaveQueryActionMenu: FC<SaveQueryActionMenuProps> = ({ saveAs }) => {
    // const handleClose = (event: MouseEvent) => {
    //     // event.stopPropagation();
    //     console.log('handle close');
    // };

    const handleSaveAs = (event: MouseEvent) => {
        event.stopPropagation();
        saveAs();
    };

    const listItemStyles = 'px-2 py-3 cursor-pointer hover:bg-neutral-light-4';

    return (
        <>
            <Popover>
                <PopoverTrigger
                    className='inline-flex items-center justify-center whitespace-nowrap rounded-3xl text-sm ring-offset-background transition-colors hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 active:no-underline bg-neutral-light-5 text-neutral-dark-0 shadow-outer-1 hover:bg-secondary hover:text-white h-9 px-4 py-1 text-xs rounded-l-none pl-2 -ml-1 '
                    onClick={(event) => event.stopPropagation()}>
                    <AppIcon.CaretDown size={10} />{' '}
                </PopoverTrigger>
                <PopoverContent className='p-0 w-28'>
                    <div className={listItemStyles} onClick={handleSaveAs}>
                        Save As
                    </div>
                </PopoverContent>
            </Popover>
        </>
    );
};

export default SaveQueryActionMenu;
