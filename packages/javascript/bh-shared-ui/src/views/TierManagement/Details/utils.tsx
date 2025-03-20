import { Skeleton } from '@bloodhoundenterprise/doodleui';
import { CSSProperties, FC } from 'react';

export const ItemSkeleton = (title: string, key: number, style?: CSSProperties) => {
    return (
        <li
            key={key}
            data-testid={`tier-management_details_${title.toLowerCase()}-list_loading-skeleton`}
            style={style}
            className='border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative h-full w-full'>
            <Skeleton className='h-10 rounded-none min-h-10' />
        </li>
    );
};

export const itemSkeletons = [ItemSkeleton, ItemSkeleton, ItemSkeleton];

const isActive = (selected: number | null, itemId: number) => {
    return selected === itemId;
};

export const SelectedHighlight: FC<{ selected: number | null; itemId: number; title: string }> = ({
    selected,
    itemId,
    title,
}) => {
    return isActive(selected, itemId) ? (
        <div
            data-testid={`tier-management_details_${title.toLowerCase()}-list_active-${title.toLowerCase()}-item-${selected}`}
            className='h-full bg-primary pr-1 absolute'></div>
    ) : null;
};
