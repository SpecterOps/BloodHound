import { FC } from 'react';
import { usePZPathParams } from '../../../hooks';

export const SelectedHighlight: FC<{
    itemId: string | number;
    type: 'tag' | 'selector' | 'member';
}> = ({ itemId, type }) => {
    const { tagId, selectorId, memberId } = usePZPathParams();

    const itemIdStr = itemId.toString();
    const activeType = memberId ? 'member' : selectorId ? 'selector' : 'tag';

    if (activeType !== type) {
        return null;
    }

    const isActive =
        (type === 'tag' && tagId === itemIdStr) ||
        (type === 'selector' && selectorId === itemIdStr) ||
        (type === 'member' && memberId === itemIdStr);

    if (!isActive) return null;

    return (
        <div
            className='h-full bg-primary pr-1 absolute'
            data-testid={`privilege-zones_details_${type}s-list_active-${type}s-item-${itemId}`}
        />
    );
};
