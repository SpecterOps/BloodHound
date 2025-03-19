import { Badge, Card } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';

type ObjectCountPanelProps = {
    data: any;
};

const ObjectCountPanel: FC<ObjectCountPanelProps> = ({ data }) => {
    return (
        <Card className='flex justify-between h-[435px] px-6 pt-6 select-none'>
            <div>Total Count</div>
            <Badge label='244' />
            {data?.map((count: any) => {
                return (
                    <div key={count}>
                        <p>{data.name}</p>
                        <Badge label={data.count} />
                    </div>
                );
            })}
        </Card>
    );
};

export default ObjectCountPanel;
