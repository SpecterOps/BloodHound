import { Card } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import LargeRightArrow from '../../../components/AppIcon/Icons/LargeRightArrow';
import Plus from '../../../components/AppIcon/Icons/Plus';
import { ROUTE_TIER_MANAGEMENT_DETAILS } from '../../../routes';
import { useAppNavigate } from '../../../utils';

type SummaryCardProps = {
    title: string;
    selectorCount: number | undefined;
    memberCount: number | undefined;
    id: number;
};

const SummaryCard: FC<SummaryCardProps> = ({ title, selectorCount, memberCount, id }) => {
    const navigate = useAppNavigate();
    return (
        <Card className='w-full flex px-6 py-4 rounded-xl'>
            <div className=' text-xl flex-1 flex items-center justify-center font-bold'>{title}</div>
            <LargeRightArrow className='w-8 h-16' />
            <div className='flex-1 flex flex-col items-center justify-center'>
                <p className='text-l font-semibold'>Selectors</p>
                <p className='text-2xl font-thin'>{selectorCount}</p>
            </div>
            <LargeRightArrow className='w-8 h-16' />
            <div className='flex-1 flex flex-col items-center justify-center'>
                <p className='text-l font-semibold'>Members</p>
                <p className='text-2xl font-thin'>{memberCount}</p>
            </div>

            <div className='flex-1 flex flex-col items-center justify-center border-l border-black dark:border-white text-sm'>
                <div
                    onClick={(e) => {
                        // Prevent event bubbling for the view details action
                        e.stopPropagation();
                        navigate(`/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}/tier/${id}`);
                    }}
                    className=' flex items-center space-x-2 hover:underline cursor-pointer'
                    role='button'
                    tabIndex={0}
                    aria-label={`View details for ${title}`}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            e.stopPropagation();
                            navigate(`/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}/tier/${id}`);
                        }
                    }}>
                    <Plus />
                    <p>View Details</p>
                </div>
            </div>
        </Card>
    );
};

export default SummaryCard;
