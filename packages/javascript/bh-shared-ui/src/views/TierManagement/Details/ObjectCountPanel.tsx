import { Badge, Card } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import { useQuery } from 'react-query';
import { apiClient } from '../../../utils';

type ObjectCountPanelProps = {
    selectedTier: number;
};

const ObjectCountPanel: FC<ObjectCountPanelProps> = ({ selectedTier }) => {
    const objectsCountQuery = useQuery(['asset-group-labels-count'], () => {
        return apiClient.getAssetGroupMembersCount(selectedTier.toString()).then((res) => {
            return res.data.data;
        });
    });

    if (objectsCountQuery.isLoading) {
        // handle loading state
        // render a skeleton
        return null;
    }
    if (objectsCountQuery.isError) {
        // display some error view
        return null;
    }

    if (objectsCountQuery.isSuccess) {
        const { total_count, counts } = objectsCountQuery.data;
        return (
            <Card className='flex justify-between h-[478px] px-6 pt-6 select-none'>
                <div>Total Count</div>
                <Badge label={total_count.toString()} />
                {Object.entries(counts).map(([key, value]) => {
                    return (
                        <div key={key}>
                            <p>{key}</p>
                            <Badge label={value.toString()} />
                        </div>
                    );
                })}
            </Card>
        );
    }
};

export default ObjectCountPanel;
