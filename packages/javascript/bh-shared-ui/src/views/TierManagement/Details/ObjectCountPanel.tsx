import { Badge, Card } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';

type ObjectCountPanelProps = {
    data: {
        total_count: number;
        counts: Record<string, number>;
    };
};

const ObjectCountPanel: FC<ObjectCountPanelProps> = (props: any) => {
    if (!props.data) {
        return null;
    }
    if (props.data.isLoading || props.data.isError) {
        return null;
    }
    if (props.data.isSuccess) {
        return (
            <Card className='flex justify-between h-[478px] px-6 pt-6 select-none'>
                <div>Total Count</div>
                <Badge label={props.data.data.total_count?.toString()} />
                {props.data.data.counts &&
                    Object.entries(props.data.data.counts).map(([key, value]) => {
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
