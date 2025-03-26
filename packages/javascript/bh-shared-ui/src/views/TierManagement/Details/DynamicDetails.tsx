import { Card } from '@bloodhoundenterprise/doodleui';
import { AssetLabel, AssetSelector } from 'js-client-library';
import { FC } from 'react';

type DynamicDetailsProps = {
    data: AssetSelector | AssetLabel | undefined;
    isCypher?: boolean;
};

const isSelector = (data: any): data is AssetSelector => {
    return 'seeds' in data;
};

const isLabel = (data: any): data is AssetLabel => {
    return 'asset_group_tier_id' in data;
};

const DynamicDetails: FC<DynamicDetailsProps> = ({ data, isCypher }) => {
    if (!data) {
        return null;
    }
    const lastUpdated = new Date(data.updated_at).toLocaleDateString();

    return (
        <Card className='h-[280px] mb-[24px] px-6 pt-6 select-none overflow-y-auto'>
            <div className='text-xl font-bold'>{data ? data.name : 'Nothing Data'}</div>
            <div className='flex flex-wrap gap-x-2'>
                {isLabel(data) && data.position !== null && (
                    <>
                        <p className='font-bold'>Tier:</p> <p>{data.position}</p>
                    </>
                )}
            </div>
            <div className='flex flex-wrap gap-x-2'>
                <p className='font-bold'>Description:</p> <p>{data.description}</p>
            </div>
            <div className='flex flex-wrap gap-x-2'>
                <p className='font-bold'>Created by:</p> <p>{data.created_by}</p>
            </div>
            <div className='flex flex-wrap gap-x-2'>
                <p className='font-bold'>Last Updated:</p> <p>{lastUpdated}</p>
            </div>
            {isLabel(data) && (
                <div className='flex flex-wrap gap-x-2'>
                    <p className='font-bold'>Certification Enabled:</p> <p>{data.require_certify}</p>
                </div>
            )}
            {isSelector(data) && (
                <>
                    <div className='flex flex-wrap gap-x-2'>
                        <p className='font-bold'>Type:</p> <p>{isCypher ? 'Cypher' : 'Object'}</p>
                    </div>
                    <div className='flex flex-wrap gap-x-2'>
                        <p className='font-bold'>Automatic Certification:</p>{' '}
                        <p>{data.auto_certify ? 'Enabled' : 'Disabled'}</p>
                    </div>
                    <div className='flex flex-wrap gap-x-2'>
                        <p className='font-bold'>Selector Enabled:</p>{' '}
                        <p>{data.disabled_at ? 'Enabled' : 'Disabled'}</p>
                    </div>
                </>
            )}
        </Card>
    );
};

export default DynamicDetails;
